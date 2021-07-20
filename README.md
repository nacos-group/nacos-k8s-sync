### 简介
该项目用于同步Kubernetes和Nacos之间的服务信息。

目前该项目仅支持 Kubernetes Service -> Nacos Service 的同步

### TODO
- ~~增加高性能zap的logger~~
- 增加 Nacos Service -> Kubernetes Service 的同步
- 监听K8s集群中的多个Namespace
- Nacos支持多Namespace注册
- 服务信息的获取方式的兜底方案，比如从Service的Spec获取
- 单元测试

### 代码提交需知
- 需要运行一下 `make precommit`，处理完imports排序后并保证编译成功，方可提交