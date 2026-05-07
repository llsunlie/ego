# ego · 后端逻辑实现

> 每个核心流程的 Go 伪代码，串联数据库与 gRPC 接口

## F0 登录/自动注册

**触发 RPC：** `Login`

```go
func (s *EgoServer) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
    account := req.Account
    password := req.Password

    // 1. 查用户是否存在
    row := db.QueryRow("SELECT id, password_hash FROM users WHERE account = $1", account)
    if row == nil {
        // 不存在 → 自动注册
        hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        userID := uuid.New().String()
        db.Exec(`
            INSERT INTO users (id, account, password_hash, created_at)
            VALUES ($1, $2, $3, $4)
        `, userID, account, hash, time.Now())

        token := generateJWT(userID)
        return &pb.LoginRes{Token: token, Created: true}, nil
    }

    // 2. 存在 → 验证密码
    userID := row["id"].(string)
    hash := row["password_hash"].(string)
    if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
        return nil, status.Error(codes.Unauthenticated, "密码错误")
    }

    token := generateJWT(userID)
    return &pb.LoginRes{Token: token, Created: false}, nil
}
```

### JWT 工具函数

```go
func generateJWT(userID string) string {
    claims := jwt.MapClaims{
        "user_id": userID,
        "iat":     time.Now().Unix(),
        "exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),  // 30 天过期
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, _ := token.SignedString(jwtSecret)
    return s
}

// 从 gRPC metadata 提取 JWT → 解析 user_id，所有接口统一调用
func extractUserID(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", status.Error(codes.Unauthenticated, "missing metadata")
    }
    authHeader := md.Get("authorization")
    if len(authHeader) == 0 || !strings.HasPrefix(authHeader[0], "Bearer ") {
        return "", status.Error(codes.Unauthenticated, "missing token")
    }
    tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")

    claims := &jwt.MapClaims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil || !token.Valid {
        return "", status.Error(codes.Unauthenticated, "invalid token")
    }
    return (*claims)["user_id"].(string), nil
}
```

---

## F1 主流程：写字 → 回声 → 观察

**触发 RPC：** `CreateMoment` + `GenerateInsight`（前端并行调用）

### CreateMoment

```go
func (s *EgoServer) CreateMoment(ctx context.Context, req *pb.CreateMomentReq) (*pb.CreateMomentRes, error) {
    userID, err := extractUserID(ctx)

    content := req.Content
    traceID := req.TraceID
    if traceID == "" {
        traceID = uuid.New().String()
    }

    // 1. 调 AI 生成 embedding
    embedding, err := s.ai.Embed(ctx, content)
    if err != nil {
        embedding = nil
    }

    // 2. 写入 Moment（显式设置所有字段，无 DEFAULT）
    moment := &Moment{
        ID:        uuid.New().String(),
        UserID:    userID,
        Content:   content,
        Embedding: embedding,
        TraceID:   traceID,
        Connected: false,
        CreatedAt: time.Now(),
    }
    db.Exec(`
        INSERT INTO moments (id, user_id, content, embedding, trace_id, connected, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, moment.ID, moment.UserID, moment.Content, moment.Embedding, moment.TraceID, moment.Connected, moment.CreatedAt)

    // 3. 向量搜索匹配回声（仅当前用户）
    var echo *pb.Echo
    if embedding != nil {
        rows := db.Query(`
            SELECT id, content, created_at, connected, trace_id,
                   1 - (embedding <=> $1) AS similarity
            FROM moments
            WHERE user_id = $2
              AND id != $3
              AND embedding IS NOT NULL
            ORDER BY embedding <=> $1
            LIMIT 4
        `, embedding, userID, moment.ID)

        if len(rows) > 0 {
            target := rows[0]
            candidates := rows[1:]
            echo = &pb.Echo{
                Id:           uuid.New().String(),
                TargetMoment: rowToMoment(target),
                Candidates:   rowsToMoments(candidates),
                Similarity:   target["similarity"].(float32),
            }
        }
    }

    return &pb.CreateMomentRes{
        Moment: momentToProto(moment),
        Echo:   echo,
    }, nil
}
```

### GenerateInsight

```go
func (s *EgoServer) GenerateInsight(ctx context.Context, req *pb.GenerateInsightReq) (*pb.GenerateInsightRes, error) {
    userID, err := extractUserID(ctx)

    // 1. 查到回声 Moment 的内容作为上下文
    var echoContent string
    if req.EchoMomentId != "" {
        row := db.QueryRow("SELECT content FROM moments WHERE id = $1 AND user_id = $2", req.EchoMomentId, userID)
        echoContent = row["content"]
    }

    // 2. 调 AI 生成观察（纯 AI 调用，不写库）
    result, err := s.ai.GenerateInsight(ctx, req.CurrentContent, echoContent)
    if err != nil {
        return nil, status.Error(codes.Unavailable, "AI service unavailable")
    }

    return &pb.GenerateInsightRes{
        Insight: &pb.Insight{
            Id:               uuid.New().String(),
            Text:             result.Text,
            RelatedMomentIds: result.RelatedMomentIDs,  // AI 返回引用的原话 id
        },
    }, nil
}
```

**时序：** 前端发 `CreateMoment` → 拿到 moment+echo → 渲染 EchoCard，同时发 `GenerateInsight` → 拿到 insight → 渲染 InsightCard。两个 RPC 独立，后者失败只隐藏卡片不报错。

---

## F2 顺着再想想：深度循环

**触发 RPC：** `CreateMoment`（同 trace_id）+ `GenerateInsight`

与 F1 完全相同的两个 RPC，唯一的区别是 `CreateMomentReq.TraceID` 传入已有的 trace_id。

```go
// CreateMoment 内部逻辑不变，trace_id 复用：
// - moments 表追加新行，trace_id 相同
// - 向量搜索时排除同 trace 下的 Moment（避免自己匹配自己）
```

向量搜索需增加过滤条件，排除当前 Trace 已有的 Moment：

```sql
SELECT id, content, created_at, connected, trace_id,
       1 - (embedding <=> $1) AS similarity
FROM moments
WHERE user_id = $user_id
  AND id != $2
  AND embedding IS NOT NULL
  AND trace_id != $current_trace_id
ORDER BY embedding <=> $1
LIMIT 4;
```

---

## F3 收进星图

**触发 RPC：** `StashTrace`

```go
func (s *EgoServer) StashTrace(ctx context.Context, req *pb.StashTraceReq) (*pb.StashTraceRes, error) {
    userID, err := extractUserID(ctx)

    traceID := req.TraceID
    x, y := req.X, req.Y

    // 1. 创建星星（显式设置所有字段）
    star := &Star{
        ID:          uuid.New().String(),
        UserID:      userID,
        TraceID:     traceID,
        X:           x,
        Y:           y,
        VisualState: "dim",
        Rhythm:      rand.Float64(),
        CreatedAt:   time.Now(),
    }
    db.Exec(`
        INSERT INTO stars (id, user_id, trace_id, x, y, visual_state, rhythm, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, star.ID, star.UserID, star.TraceID, star.X, star.Y, star.VisualState, star.Rhythm, star.CreatedAt)

    // 2. 异步触发聚类（goroutine，不阻塞响应）
    go s.clusterStar(star)

    return &pb.StashTraceRes{
        Star: starToProto(star),
    }, nil
}

// 异步聚类逻辑
func (s *EgoServer) clusterStar(star *Star) {
    // 1. 获取该 Trace 下所有 Moment 的 embedding 均值
    rows := db.Query(`
        SELECT embedding FROM moments
        WHERE user_id = $1 AND trace_id = $2 AND embedding IS NOT NULL
    `, star.UserID, star.TraceID)
    if len(rows) == 0 {
        return
    }
    starVec := avgEmbeddings(rows)

    // 2. 与同一用户已有星座做相似度匹配
    constellations := db.Query("SELECT * FROM constellations WHERE user_id = $1", star.UserID)
    var bestMatch *Constellation
    var bestScore float64
    for _, c := range constellations {
        cVec := getConstellationVector(c.ID, star.UserID)
        score := cosineSimilarity(starVec, cVec)
        if score > bestScore && score > 0.75 {
            bestScore = score
            bestMatch = c
        }
    }

    // 3. 根据匹配结果更新星座
    if bestMatch != nil {
        db.Exec("INSERT INTO constellation_stars (constellation_id, star_id) VALUES ($1, $2)",
            bestMatch.ID, star.ID)

        newStatus := evaluateConstellationStatus(bestMatch.ID)
        db.Exec("UPDATE constellations SET status = $1, updated_at = $2 WHERE id = $3",
            newStatus, time.Now(), bestMatch.ID)

        if newStatus == "formed" {
            go s.regenerateConstellationAssets(bestMatch.ID, star.UserID)
        }

        // 更新该星座下所有 Moment 的 connected 标记
        db.Exec(`
            UPDATE moments SET connected = TRUE
            WHERE user_id = $1 AND trace_id IN (
                SELECT trace_id FROM stars s
                JOIN constellation_stars cs ON cs.star_id = s.id
                WHERE s.user_id = $1 AND cs.constellation_id = $2
            )
        `, star.UserID, bestMatch.ID)
    } else {
        // 无匹配 → 新建孤星星座
        c := &Constellation{
            ID:        uuid.New().String(),
            UserID:    star.UserID,
            Name:      "未命名的星",
            Status:    "lone",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }
        db.Exec(`
            INSERT INTO constellations (id, user_id, name, status, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, c.ID, c.UserID, c.Name, c.Status, c.CreatedAt, c.UpdatedAt)
        db.Exec("INSERT INTO constellation_stars (constellation_id, star_id) VALUES ($1, $2)",
            c.ID, star.ID)
    }
}

func (s *EgoServer) regenerateConstellationAssets(constellationID string, userID string) {
    // 收集该星座所有 Moment
    moments := db.Query(`
        SELECT m.* FROM moments m
        JOIN stars s ON s.trace_id = m.trace_id AND s.user_id = $2
        JOIN constellation_stars cs ON cs.star_id = s.id
        WHERE m.user_id = $2 AND cs.constellation_id = $1
    `, constellationID, userID)

    // 1. AI 生成星座名称
    name := s.ai.GenerateConstellationName(moments)
    db.Exec("UPDATE constellations SET name = $1, updated_at = $3 WHERE id = $2", name, constellationID, time.Now())

    // 2. AI 生成观察
    insight := s.ai.GenerateConstellationInsight(moments)
    db.Exec("DELETE FROM insights WHERE constellation_id = $1 AND user_id = $2", constellationID, userID)
    db.Exec(`
        INSERT INTO insights (id, user_id, constellation_id, text, related_moment_ids, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, uuid.New().String(), userID, constellationID, insight.Text, insight.RelatedIDs, time.Now())

    // 3. AI 生成 past_self_cards（最多 3 张）
    db.Exec("DELETE FROM past_self_cards WHERE constellation_id = $1 AND user_id = $2", constellationID, userID)
    cards := s.ai.GeneratePastSelfCards(moments)
    for _, card := range cards {
        db.Exec(`
            INSERT INTO past_self_cards (id, user_id, constellation_id, title, opening_line, context_moment_ids, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
        `, uuid.New().String(), userID, constellationID, card.Title, card.OpeningLine, card.ContextMomentIDs, time.Now())
    }

    // 4. AI 生成 topic_prompts（3 个）
    db.Exec("DELETE FROM topic_prompts WHERE constellation_id = $1 AND user_id = $2", constellationID, userID)
    prompts := s.ai.GenerateTopicPrompts(moments)
    for _, p := range prompts {
        db.Exec(`
            INSERT INTO topic_prompts (id, user_id, constellation_id, anchor_moment_id, question, created_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, uuid.New().String(), userID, constellationID, p.AnchorMomentID, p.Question, time.Now())
    }
}
```

---

## F4 记忆光点：盲盒

**触发 RPC：** `GetRandomMoments`

```go
func (s *EgoServer) GetRandomMoments(ctx context.Context, req *pb.GetRandomMomentsReq) (*pb.GetRandomMomentsRes, error) {
    userID, err := extractUserID(ctx)

    count := req.Count
    if count == 0 {
        count = 3
    }

    rows := db.Query(`
        SELECT id, content, created_at, connected, trace_id
        FROM moments
        WHERE user_id = $1 AND embedding IS NOT NULL
        ORDER BY random()
        LIMIT $2
    `, userID, count)

    return &pb.GetRandomMomentsRes{
        Moments: rowsToMoments(rows),
    }, nil
}
```

**要点：**
- `ORDER BY random()` 在小数据量下性能可接受（<10K 行）。数据量大后改用 `TABLESAMPLE` 或预计算随机偏移
- `WHERE embedding IS NOT NULL` 排除刚写入尚未生成 embedding 的 Moment

---

## F5 过往时间线

**触发 RPC：** `ListMoments`

```go
func (s *EgoServer) ListMoments(ctx context.Context, req *pb.ListMomentsReq) (*pb.ListMomentsRes, error) {
    userID, err := extractUserID(ctx)

    pageSize := req.PageSize
    if pageSize == 0 {
        pageSize = 20
    }

    var rows []Row
    if req.Cursor == "" {
        rows = db.Query(`
            SELECT id, content, created_at, connected, trace_id
            FROM moments
            WHERE user_id = $1
            ORDER BY created_at DESC
            LIMIT $2
        `, userID, pageSize+1)
    } else {
        rows = db.Query(`
            SELECT id, content, created_at, connected, trace_id
            FROM moments
            WHERE user_id = $1
              AND created_at < (SELECT created_at FROM moments WHERE id = $2 AND user_id = $1)
            ORDER BY created_at DESC
            LIMIT $3
        `, userID, req.Cursor, pageSize+1)
    }

    hasMore := len(rows) > pageSize
    if hasMore {
        rows = rows[:pageSize]
    }

    // 前端自行按月分组，后端只返回扁平列表
    var nextCursor string
    if hasMore && len(rows) > 0 {
        nextCursor = rows[len(rows)-1]["id"].(string)
    }

    return &pb.ListMomentsRes{
        Moments:    rowsToMoments(rows),
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }, nil
}
```

**说明：** 前端根据 `created_at` 自行按月分组。`connected` 字段决定是否显示 "✦ 已联结" 标记。

---

## F6 星图探索

### ListConstellations

```go
func (s *EgoServer) ListConstellations(ctx context.Context, req *pb.ListConstellationsReq) (*pb.ListConstellationsRes, error) {
    userID, err := extractUserID(ctx)

    rows := db.Query(`
        SELECT c.id, c.name, c.status, c.updated_at,
               s.id AS star_id, s.trace_id, s.x, s.y, s.visual_state, s.rhythm,
               (SELECT COUNT(*) FROM moments m WHERE m.trace_id = s.trace_id AND m.user_id = $1) AS moment_count
        FROM constellations c
        JOIN constellation_stars cs ON cs.constellation_id = c.id
        JOIN stars s ON s.id = cs.star_id
        WHERE c.user_id = $1
        ORDER BY c.updated_at DESC
    `, userID)

    // 内存组装：group by constellation_id，moment_count 按星座汇总
    result := groupByConstellation(rows)

    var totalStars int
    for _, c := range result {
        totalStars += len(c.Stars)
    }

    return &pb.ListConstellationsRes{
        Constellations: result,
        TotalStarCount: int32(totalStars),
    }, nil
}
```

### GetConstellation

```go
func (s *EgoServer) GetConstellation(ctx context.Context, req *pb.GetConstellationReq) (*pb.GetConstellationRes, error) {
    userID, err := extractUserID(ctx)

    cid := req.ConstellationID

    var (
        constellation  *Constellation
        insight        *Insight
        pastSelfCards  []PastSelfCard
        topicPrompts   []TopicPrompt
        moments        []Moment
    )

    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        constellation = db.QueryRow("SELECT * FROM constellations WHERE id = $1 AND user_id = $2", cid, userID)
        return nil
    })
    g.Go(func() error {
        insight = db.QueryRow("SELECT * FROM insights WHERE constellation_id = $1 AND user_id = $2", cid, userID)
        return nil
    })
    g.Go(func() error {
        pastSelfCards = db.Query("SELECT * FROM past_self_cards WHERE constellation_id = $1 AND user_id = $2", cid, userID)
        return nil
    })
    g.Go(func() error {
        topicPrompts = db.Query("SELECT * FROM topic_prompts WHERE constellation_id = $1 AND user_id = $2", cid, userID)
        return nil
    })

    g.Wait()

    // 查询该星座下全部原话
    allMoments := db.Query(`
        SELECT m.* FROM moments m
        JOIN stars s ON s.trace_id = m.trace_id AND s.user_id = $2
        JOIN constellation_stars cs ON cs.star_id = s.id
        WHERE m.user_id = $2 AND cs.constellation_id = $1
        ORDER BY m.created_at DESC
    `, cid, userID)

    refMomentIDs := collectReferencedMomentIDs(insight, pastSelfCards, topicPrompts)
    refMoments := db.Query("SELECT * FROM moments WHERE id = ANY($1) AND user_id = $2", refMomentIDs, userID)
    refMap := toMap(refMoments)

    // 组装响应
    return &pb.GetConstellationRes{
        Constellation: constellationToProto(constellation),
        Insight:       insightToProto(insight),
        PastSelfCards: pastSelfCardsToProto(pastSelfCards, refMap),
        TopicPrompts:  topicPromptsToProto(topicPrompts, refMap),
        Moments:       rowsToMoments(allMoments),
    }, nil
}
```

**说明：** 6 条查询并行执行（Go errgroup），全量原话和引用原话分开取。前端拿到后纯渲染。

---

## F7 对话模式

### StartChat

```go
func (s *EgoServer) StartChat(ctx context.Context, req *pb.StartChatReq) (*pb.StartChatRes, error) {
    userID, err := extractUserID(ctx)

    // 恢复旧会话
    if req.ChatSessionID != "" {
        messages := db.Query(`
            SELECT * FROM chat_messages
            WHERE session_id = $1 AND user_id = $2 ORDER BY created_at
        `, req.ChatSessionID, userID)

        if len(messages) > 0 {
            return &pb.StartChatRes{
                ChatSessionId: req.ChatSessionID,
                Opening:       messages[len(messages)-1].ToProto(),
                History:       messagesToProto(messages),
            }, nil
        }
    }

    // 新建会话
    session := &ChatSession{
        ID:               uuid.New().String(),
        UserID:           userID,
        PastSelfCardID:   req.PastSelfCardID,
        ContextMomentIDs: req.ContextMomentIDs,
        CreatedAt:        time.Now(),
    }
    db.Exec(`
        INSERT INTO chat_sessions (id, user_id, past_self_card_id, context_moment_ids, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, session.ID, session.UserID, session.PastSelfCardID, session.ContextMomentIDs, session.CreatedAt)

    // 加载上下文原话
    contextMoments := db.Query(`
        SELECT * FROM moments WHERE id = ANY($1) AND user_id = $2
    `, req.ContextMomentIDs, userID)

    // AI 生成开场白
    opening := s.ai.GenerateOpeningLine(ctx, contextMoments)
    openingMsg := &ChatMessage{
        ID:        uuid.New().String(),
        UserID:    userID,
        SessionID: session.ID,
        Role:      "past_self",
        Content:   opening,
        Timestamp: time.Now(),
    }
    db.Exec(`
        INSERT INTO chat_messages (id, user_id, session_id, role, content, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, openingMsg.ID, openingMsg.UserID, openingMsg.SessionID, openingMsg.Role, openingMsg.Content, openingMsg.Timestamp)

    return &pb.StartChatRes{
        ChatSessionId: session.ID,
        Opening:       openingMsg.ToProto(),
        History:       nil,
    }, nil
}
```

### SendMessage

```go
func (s *EgoServer) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageRes, error) {
    userID, err := extractUserID(ctx)

    // 1. 存用户消息
    userMsg := &ChatMessage{
        ID:        uuid.New().String(),
        UserID:    userID,
        SessionID: req.ChatSessionID,
        Role:      "user",
        Content:   req.Content,
        Timestamp: time.Now(),
    }
    db.Exec(`
        INSERT INTO chat_messages (id, user_id, session_id, role, content, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, userMsg.ID, userMsg.UserID, userMsg.SessionID, userMsg.Role, userMsg.Content, userMsg.Timestamp)

    // 2. 加载会话上下文
    session := db.QueryRow("SELECT * FROM chat_sessions WHERE id = $1 AND user_id = $2", req.ChatSessionID, userID)
    contextMoments := db.Query(`
        SELECT * FROM moments WHERE id = ANY($1) AND user_id = $2
    `, session.ContextMomentIDs, userID)
    history := db.Query(`
        SELECT * FROM chat_messages WHERE session_id = $1 AND user_id = $2 ORDER BY created_at
    `, req.ChatSessionID, userID)

    // 3. AI 以 past_self 身份回复
    reply, err := s.ai.ChatReply(ctx, contextMoments, history, req.Content)
    if err != nil {
        return nil, status.Error(codes.Unavailable, "this version of you is thinking...")
    }

    // 4. 存 AI 回复
    replyMsg := &ChatMessage{
        ID:                uuid.New().String(),
        UserID:            userID,
        SessionID:         req.ChatSessionID,
        Role:              "past_self",
        Content:           reply.Text,
        ReferencedMoments: reply.References,
        Timestamp:         time.Now(),
    }
    db.Exec(`
        INSERT INTO chat_messages (id, user_id, session_id, role, content, referenced_moments, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, replyMsg.ID, replyMsg.UserID, replyMsg.SessionID, replyMsg.Role, replyMsg.Content, replyMsg.ReferencedMoments, replyMsg.Timestamp)

    return &pb.SendMessageRes{
        Reply: replyMsg.ToProto(),
    }, nil
}
```

**AI 约束（Prompt 层面保证，非代码逻辑）：**
- 原话是地基，不可伪造
- 可反问、可自省，不可捏造
- 超出范围回复："这个我那时候没想过，要不要你说给我听？"
- 每条回复标注 `referenced_moments`

---

## F8 话题引子：从星图回到此刻

**后端无新 RPC。** `GetConstellation` 已在 F6 覆盖，话题引子数据随详情页一起返回。

前端点击话题引子后的行为纯客户端：
```
Navigator.pop() → switch to NowPage → 打开写字区，CreateMomentReq.Topic = prompt.Question
```

`CreateMomentReq.topic` 字段由前端随请求传入（话题引子文本），后端当前版本不持久化到 moments 表，预留用于未来洞察关联分析。

---

## F9 冷启动

**无特殊后端逻辑。** 所有 RPC 正常处理用户空数据：

| RPC | 空状态返回 |
|-----|-----------|
| `Login` | 首次登录自动注册，`created: true` |
| `GetRandomMoments` | `moments: []` 空列表 |
| `CreateMoment` | `echo: nil`（无历史可匹配） |
| `GenerateInsight` | 此 RPC 前端在 echo 为 nil 时不调用 |
| `StashTrace` | 正常写入第一颗星，聚类结果为 lone |
| `ListMoments` | `moments: []`, `has_more: false` |
| `ListConstellations` | `constellations: []`, `total_star_count: 0` |

**不需要**专门的"初始化用户"逻辑——用户 Login 时自动注册，第一条 Moment 写入时才产生第一行业务数据。

---

