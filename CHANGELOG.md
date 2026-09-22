# Changelog

本文件将记录本项目的所有重要变更。

该格式基于 [Keep a Changelog](https://keepachangelog.com),
本项目遵循 [语义化版本控制](https://semver.org/spec/v2.0.0.html)。

## [v1.2.0] - 2026-09-22

### ✨ Feat

- 增加安全相关内容
    - 有符号整数安全运算, 防止余额/累计值加减溢出回绕
    - 用户产生内容的 HTML 消毒(纵深防御, 剥离常见 XSS 向量)

## [v1.1.2] - 2026-09-19

### 🐞 Fix

- EncodeURLComponent 按 JS encodeURIComponent 语义对字符串 s 做百分号编码.

## [v1.1.1] - 2026-09-16

### 📦 Build

- 依赖升级

## [v1.1.0] - 2026-09-02

### ✨ Feat

- 增加 domain 域名相关的方法

### 📦 Build

- 依赖升级

## [v1.0.2] - 2026-08-12

### 📦 Build

- 依赖升级

## [v1.0.1] - 2026-07-31

### ✨ Feat

- 用标准库 PBKDF2 替换 bcrypt 密码哈希实现

## [v1.0.0] - 2026-07-31

### 🐞 Fix

- 修复 lint gocognit

### 📦 Build

- 依赖升级

## [v0.16.0] - 2026-05-16

### ♻️ Refactor

- SQLTemplate 重命名为 UnsafeSQLTemplate，补充安全警告

### ✨ Feat

- 新增 boot 公共引导包 子进程启动、日志尾缓存、端口占用检测
- GORM logger 支持 ErrorClassifier，按错误类型调整日志级别

### 🐞 Fix

- NOGROUP pending 循环降为 Debug + 计数限频，避免日志风暴
- EncryptAES/DecryptAES 未传 IV 时升级为 AES-GCM，兼容旧格式解密

### 🔧 Chore

- lint 修复

## [v0.15.0] - 2026-05-13

### ✨ Feat

- 新增计数器原子初始化能力
- 通用雪花 ID 生成器封装保证并发唯一

### 🐞 Fix

- 修复高并发下计数器事务冲突导致的限流异常
- lint 警告

### 📦 Build

- 依赖升级
- 1.26.3

## [v0.14.3] - 2026-04-28

### ♻️ Refactor

- 降低认知复杂度

### 🔧 Chore

- 更新依赖

## [v0.14.2] - 2026-04-26

### 🐞 Fix

- **markdown:** 行内代码包裹的 pay-* 起标签不再吞掉后续内容

## [v0.14.1] - 2026-04-23

### 🐞 Fix

- 脱密字段非 string 不使用 panic

### 🔧 Chore

- 依赖升级

## [v0.14.0] - 2026-04-19

### ✨ Feat

- pay-video 增加 has-material 的属性

## [v0.13.2] - 2026-04-12

### 🐞 Fix

- 敏感字段判断优先兼容 JSON tag, 避免 garble 混淆字段名后失效.

### 🔧 Chore

- go 1.26.2

## [v0.13.1] - 2026-03-29

### 🐞 Fix

- **req:** 请求增加 server 版本信息修改 option 模式

## [v0.13.0] - 2026-03-29

### ✨ Feat

- **req:** 请求增加 server 版本信息

## [v0.12.6] - 2026-03-28

### ⚡️ Perf

- 减少不必要的流程提升日志打印的性能
- 优化日志界别减少生产输出提升性能

### 🐞 Fix

- 默认增加验证码和健壮一些逻辑

## [v0.12.5] - 2026-03-19

### ⏪ Revert

- 检验器仅兼容指针不改原有功能 正整数

## [v0.12.4] - 2026-03-18

### 🐞 Fix

- 校验器兼容指针类型
- 修复隐式 TLS 连接池复用失效连接导致发信异常

## [v0.12.3] - 2026-03-17

### 🐞 Fix

- 反引号代码块和付费组件嵌套的问题

## [v0.12.2] - 2026-03-13

### 🐞 Fix

- 依赖带来的用法变更

### 📦 Build

- 更新依赖

## [v0.12.1] - 2026-03-01

### 🐞 Fix

- 敏感数据掩码处理

## [v0.12.0] - 2026-03-01

### ✨ Feat

- **pay:** 处理支付宝余额不足

## [v0.11.0] - 2026-02-27

### ✨ Feat

- **email:** 增加 SendWithBase 邮件 html 模版复用

### 📦 Build

- lint and format
- **email:** lint 降低认知复杂度

## [v0.10.0] - 2026-02-12

### 💥 Boom

- **email:** 将 email 工具提到 email 包

## [v0.9.1] - 2026-02-11

### 🐞 Fix

- 交易流水状态校验

## [v0.9.0] - 2026-02-11

### ✨ Feat

- **model:** 交易流水类型

## [v0.8.1] - 2026-02-04

### 🐞 Fix

- **model:** 完善货币符号的 UnmarshalJSON

## [v0.8.0] - 2026-02-02

### 💥 Boom

- **model:** 自定义类型序列化为 value label

## [v0.7.0] - 2026-02-01

### ✨ Feat

- **res:** 增加 html 的响应封装

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

## [v0.5.0] - 2026-01-29

### 💥 Boom

- **cert:** 将 cert 单独构建一个repo

### 📦 Build

- 将 skill 抽离到工作区共享

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

## [v0.2.2] - 2026-01-18

### Fix

- 签名选项增加请求设定的算法

## [v0.2.1] - 2026-01-18

### Fix

- 修复一些已知问题

## [v0.2.0] - 2026-01-13

### Add

- 增加类型相关和 json 加解密的工具

## [v0.1.0] - 2026-01-12

### Add

- 工具库首发。
