# timeline Clean State Checklist

提交代码前必须通过的局部自查：

- [ ] `go test ./internal/timeline/...` 全部通过
- [ ] `go build ./...` 无编译错误
- [ ] domain 层只定义只读端口，无 proto、sqlc、pgx、platform 依赖
- [ ] app 层只依赖 domain ports，无 adapter 依赖
- [ ] `module.go` 只做模块级装配：创建 read adapter、app use case、handler，不承载业务算法或默认值
- [ ] `bootstrap/timeline.go` 只注入 DB，不直接 new writingpostgres repo 或 timeline use case
- [ ] Timeline 不读取配置、不创建 DB pool、不初始化外部 SDK
- [ ] adapter/grpc 中 proto 转换在 mapper 完成，handler 不直接操作 pb 类型
- [ ] 所有 gRPC handler 从 ctx 获取 user_id，不自行解析 JWT
- [ ] Timeline 只读，不写入 traces、moments、echos、insights 等 Writing 拥有的表
- [ ] Timeline 不直接访问 starmap 或 chat 模块的内部实现
- [ ] progress.md 已记录本次变更
