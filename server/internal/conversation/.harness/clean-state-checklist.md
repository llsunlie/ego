# conversation Clean State Checklist

提交代码前必须通过的局部自查：

- [ ] `go test ./internal/conversation/...` 全部通过
- [ ] `go build ./...` 无编译错误
- [ ] domain 层无 proto、sqlc、pgx、platform 依赖
- [ ] app 层只依赖 domain ports，无 adapter 依赖
- [ ] DefaultChatGenerator 业务策略位于 `app/`，不写在 `module.go` 或 `bootstrap/`
- [ ] `module.go` 只做模块级装配：创建 adapter、app use case、handler，不承载业务算法
- [ ] `bootstrap/chat.go` 只注入 DB，不直接 new conversation repo/usecase/policy
- [ ] Conversation 不读取配置、不创建 DB pool、不初始化外部 SDK
- [ ] adapter/grpc 中 proto 转换在 mapper 完成，handler 不直接操作 pb 类型
- [ ] 所有 gRPC handler 从 ctx 获取 user_id，不自行解析 JWT
- [ ] Conversation 不写入 stars、constellations、moments 等其他模块拥有的表
- [ ] AI 回复的引用来源在 app 层校验
- [ ] progress.md 已记录本次变更
