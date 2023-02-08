# monotonic read consistency 单调读一致性
这里是 monotonic read consistency 单调读一致性的简单demo

## sticky to a replica
这里主要以sticky to a replica 这种方式来实现单调读一致性

将同一用户的读固定到某个replica，以确保同一用户读到的数据不会回滚

现在使用Headless Service 创建的mysql集群，直接获取mysql子节点的信息比较麻烦，先直接写到配置文件了。

# 测试结果
由于集群只有一个节点，mysql的主从同步并没有复现回滚这种错误。