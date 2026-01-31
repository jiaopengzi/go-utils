# Changelog

本文件将记录本项目的所有重要变更。

该格式基于 [Keep a Changelog](https://keepachangelog.com),
本项目遵循 [语义化版本控制](https://semver.org/spec/v2.0.0.html)。

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

[v0.6.0]: https://github.com/jiaopengzi/go-utils/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/jiaopengzi/go-utils/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/jiaopengzi/go-utils/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/jiaopengzi/go-utils/compare/v0.2.2...v0.3.0
[v0.2.2]: https://github.com/jiaopengzi/go-utils/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/jiaopengzi/go-utils/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/jiaopengzi/go-utils/compare/v0.1.0...v0.2.0
