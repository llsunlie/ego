# writing Clean State Checklist

提交代码前必须通过的局部自查：

- [ ] `go test ./internal/writing/...` 全部通过（覆盖率不做硬性要求）
- [ ] `go build ./...` 无编译错误
- [ ] `./smoke.sh` 端到端测试全部 PASS
- [ ] Moment/Trace/Echo/Insight 写入逻辑仅在 Writing 模块内，未泄漏到其他模块
- [ ] 未写入 Star、Constellation、ChatSession 等其他模块拥有的表
- [ ] domain 层无 proto、sqlc、pgx、platform 依赖
- [ ] app 层无 adapter 依赖
- [ ] Echo/Insight 等 Writing 业务策略位于 `app/`，不写在 `module.go` 或 `bootstrap/`
- [ ] `module.go` 只做模块级装配：创建本模块 adapter、app use case、handler，不承载业务算法或文案
- [ ] `bootstrap/writing.go` 只注入进程级资源与外部基础能力（如 DB、EmbeddingGenerator），不直接 new Writing repo/usecase
- [ ] Writing 不读取配置、不创建 DB pool、不初始化外部 SDK、不直接创建 `platform/ai` 具体对象
- [ ] adapter/grpc 中 proto 转换在 mapper 完成，handler 不直接操作 pb 类型
- [ ] 所有 gRPC handler 从 ctx 获取 user_id，不自行解析 JWT
- [ ] Moment.Embeddings 不通过 API 暴露给客户端
- [ ] progress.md 已记录本次变更
