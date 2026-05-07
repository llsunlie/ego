# ARCHITECTURE.md - 内部设计说明

> 注意：以下为模板内容，当前模块还未进行方法架构设计

## 1. 领域模型 (Domain Model)
- **聚合根 (Aggregate Root)**: [名称占位] - [核心职责说明]
- **实体 (Entity)**: [名称占位]
- **值对象 (Value Object)**: [名称占位]

## 2. 核心领域逻辑
[描述核心算法、校验规则或领域服务逻辑的占位说明。]

## 3. 分层职责
- **domain/**: 存放上述模型及 Repository 接口定义。
- **app/**: 存放 UseCase 编排逻辑，协调领域层与基础设施。
- **adapter/**: 
  - `grpc/`: 负责将 Proto 请求转换为内部 Command/Query。
  - `postgres/`: 负责具体的持久化实现。

## 4. 依赖关系图

- [本模块] ➔ 依赖 [其他模块] 的 CONTRACT.md
- [本模块] ➔ 依赖 platform 的抽象 Port