# Changelog

本文件将记录本项目的所有重要变更。

该格式基于 [Keep a Changelog](https://keepachangelog.com),
本项目遵循 [语义化版本控制](https://semver.org/spec/v2.0.0.html)。

<a name="v0.13.1"></a>

## [v0.13.1] - 2026-03-29

### 🐞 Fix

- **req:** 请求增加 server 版本信息修改 option 模式

<a name="v0.13.0"></a>

## [v0.13.0] - 2026-03-29

### ✨ Feat

- **req:** 请求增加 server 版本信息

<a name="v0.12.6"></a>

## [v0.12.6] - 2026-03-28

### ⚡️ Perf

- 减少不必要的流程提升日志打印的性能
- 优化日志界别减少生产输出提升性能

### 🐞 Fix

- 默认增加验证码和健壮一些逻辑
<a name="v0.12.5"></a>

## [v0.12.5] - 2026-03-19

### ⏪ Revert

- 检验器仅兼容指针不改原有功能 正整数

<a name="v0.12.4"></a>

## [v0.12.4] - 2026-03-18

### 🐞 Fix

- 校验器兼容指针类型
- 修复隐式 TLS 连接池复用失效连接导致发信异常

<a name="v0.12.3"></a>

## [v0.12.3] - 2026-03-17

### 🐞 Fix

- 反引号代码块和付费组件嵌套的问题
<a name="v0.12.2"></a>

## [v0.12.2] - 2026-03-13

### 🐞 Fix

- 依赖带来的用法变更

### 📦 Build

- 更新依赖

<a name="v0.12.1"></a>

## [v0.12.1] - 2026-03-01

### 🐞 Fix

- 敏感数据掩码处理

<a name="v0.12.0"></a>

## [v0.12.0] - 2026-03-01

### ✨ Feat

- **pay:** 处理支付宝余额不足

<a name="v0.11.0"></a>

## [v0.11.0] - 2026-02-27

### ✨ Feat

- **email:** 增加 SendWithBase 邮件 html 模版复用

### 📦 Build

- lint and format
- **email:** lint 降低认知复杂度

<a name="v0.10.0"></a>

## [v0.10.0] - 2026-02-12

### 💥 Boom

- **email:** 将 email 工具提到 email 包

<a name="v0.9.1"></a>

## [v0.9.1] - 2026-02-11

### 🐞 Fix

- 交易流水状态校验

<a name="v0.9.0"></a>

## [v0.9.0] - 2026-02-11

### ✨ Feat

- **model:** 交易流水类型

<a name="v0.8.1"></a>

## [v0.8.1] - 2026-02-04

### 🐞 Fix

- **model:** 完善货币符号的 UnmarshalJSON

<a name="v0.8.0"></a>

## [v0.8.0] - 2026-02-02

### 💥 Boom

- **model:** 自定义类型序列化为 value label

<a name="v0.7.0"></a>

## [v0.7.0] - 2026-02-01

### ✨ Feat

- **res:** 增加 html 的响应封装

<a name="v0.6.0"></a>

## [v0.6.0] - 2026-01-29

### ✨ Feat

- 增加校验函数 ValidateCSR
- **utils:** 增加 list 工具函数
- **utils:** byte 工具函数

### 🐞 Fix

- 修复已知 bug

### 💥 Boom

- **req:** 将对称加密修改为非对称加密
- **res:** Cert 规范为 UserCert
- **utils:** SignOption 更新为更通用的 HAOption

<a name="v0.5.0"></a>

## [v0.5.0] - 2026-01-29

### 💥 Boom

- **cert:** 将 cert 单独构建一个repo

### 📦 Build

- 将 skill 抽离到工作区共享

<a name="v0.4.0"></a>

## [v0.4.0] - 2026-01-28

### ♻️ Refactor

- **cert:** 抽离工具到 utils 模块, 添加文件头
- **utils:** 加密相关统一命名为 crypto

### ✨ Feat

- **cert:** 封装易用函数
- **cert:** 增加证书相关功能
- **model:** 实现 AmountFenToYuan 方法
- **utils:** 字符串相关的工具函数

### 💥

- **utils:** CheckGormRowsAffected 更通用命名

### 🔧 Chore

- add res rule skill

<a name="v0.3.0"></a>

## [v0.3.0] - 2026-01-25

### Feat

- 增加分布式锁获取失败错误

### ♻️ Refactor

- code msg 实现重构
- 完善注释

### ⚡️ Perf

- 移除一些 zap 日志点

### ✨ Feat

- 增加最小消费金额函数
- 增加消息展示
- ValidateCurrency

### 🐞 Fix

- 序列化使用指针
- **utils:** 值接收器用作 Getter 指针接收器用作 Setter

### 📝 Docs

- add .markdownlint.yaml
- 使用 git-chglog

<a name="v0.2.2"></a>

## [v0.2.2] - 2026-01-18

### Fix

- 签名选项增加请求设定的算法

<a name="v0.2.1"></a>

## [v0.2.1] - 2026-01-18

### Fix

- 修复一些已知问题

<a name="v0.2.0"></a>

## [v0.2.0] - 2026-01-13

### Add

- 增加类型相关和 json 加解密的工具

<a name="v0.1.0"></a>

## [v0.1.0] - 2026-01-12

### Add

- 工具库首发。

[v0.13.1]: https://github.com/jiaopengzi/go-utils/compare/v0.13.0...v0.13.1
[v0.13.0]: https://github.com/jiaopengzi/go-utils/compare/v0.12.6...v0.13.0
[v0.12.6]: https://github.com/jiaopengzi/go-utils/compare/v0.12.5...v0.12.6
[v0.12.5]: https://github.com/jiaopengzi/go-utils/compare/v0.12.4...v0.12.5
[v0.12.4]: https://github.com/jiaopengzi/go-utils/compare/v0.12.3...v0.12.4
[v0.12.3]: https://github.com/jiaopengzi/go-utils/compare/v0.12.2...v0.12.3
[v0.12.2]: https://github.com/jiaopengzi/go-utils/compare/v0.12.1...v0.12.2
[v0.12.1]: https://github.com/jiaopengzi/go-utils/compare/v0.12.0...v0.12.1
[v0.12.0]: https://github.com/jiaopengzi/go-utils/compare/v0.11.0...v0.12.0
[v0.11.0]: https://github.com/jiaopengzi/go-utils/compare/v0.10.0...v0.11.0
[v0.10.0]: https://github.com/jiaopengzi/go-utils/compare/v0.9.1...v0.10.0
[v0.9.1]: https://github.com/jiaopengzi/go-utils/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/jiaopengzi/go-utils/compare/v0.8.1...v0.9.0
[v0.8.1]: https://github.com/jiaopengzi/go-utils/compare/v0.8.0...v0.8.1
[v0.8.0]: https://github.com/jiaopengzi/go-utils/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/jiaopengzi/go-utils/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/jiaopengzi/go-utils/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/jiaopengzi/go-utils/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/jiaopengzi/go-utils/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/jiaopengzi/go-utils/compare/v0.2.2...v0.3.0
[v0.2.2]: https://github.com/jiaopengzi/go-utils/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/jiaopengzi/go-utils/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/jiaopengzi/go-utils/compare/v0.1.0...v0.2.0
