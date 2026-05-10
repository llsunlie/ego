# identity Clean State Checklist

提交代码前必须通过的局部自查：

- [ ] `go test ./internal/identity/...` 全部通过
- [ ] `go build ./...` 无编译错误
- [ ] domain 层无 proto、sqlc、pgx、platform 依赖
- [ ] app 层只依赖 domain ports，无 adapter 依赖
- [ ] `module.go` 只做模块级装配：创建 adapter、app use case、handler，不承载业务算法
- [ ] `bootstrap/identity.go` 只注入 DB + Hasher + Tokens，不直接 new identity repo/usecase
- [ ] Identity 不读取配置、不创建 DB pool、不初始化外部 SDK
- [ ] Identity 不写入其他模块拥有的表
- [ ] `users` 是 Identity 唯一拥有的写入表
- [ ] progress.md 已记录本次变更
