v1.0 版本计划
---

> 如果你有什么其它好的点子，请在[Github](https://github.com/alibaba/kt-connect/issues/new?assignees=&labels=&template=feature_request.md&title=)中反馈

以下计划内容会随项目进展，随时进行调整：

#### v0.1.x

* 支持全局配置，简化命令行参数
* `connect`命令的vpn模式支持自动安装sshuttle依赖

#### v0.2.x

* 支持WireGuard协议连接，进一步优化Windows体验
* `exchange`和`mesh`命令使用Service名作为目标
* 支持对StatefulSet资源的`exchange`和`mesh`操作

#### v0.3.x

* 支持后台运行
* `mesh`命令支持自动设置流量规则（配合Istio）
* 支持全局参数和子命令参数任意顺序混用
