# data-demo
demo of data problem

此库用于展示一些数据问题的常用做法，如果大家有更好的解决方案，欢迎提pr。

## 环境安装
- 前置需求
k8s，helm3
### mysql集群安装
```sh
$ helm3 --namespace mysql-cluster install mysql  -f ./helm/mysql-values-replication.yaml
```
