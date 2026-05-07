# CONTRACT.md - 模块对外契约

## 1. 对外暴露的接口 (Ports)
[列出其他模块可以直接调用的 Interface 或服务入口]
- `ServiceX`: [功能描述占位]

## 2. 共享只读模型 (Read Models)
[描述其他模块通过查询接口可以获取的数据结构]

## 3. 领域事件 (Domain Events)
[描述本模块会向外部 EventBus 发布的消息]
- `[EventName]`: 触发时机及携带的 Data 载体。

## 4. 禁止事项
- 其他模块禁止直接访问本模块的 `internal/adapter/` 目录。
- 其他模块禁止直接写入本模块所属的数据库表。