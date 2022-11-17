# read-after-write consistency 写后读一致性
这里是 read-after-write consistency 写后读一致性的简单demo

## Pinning User to Master
这里主要以Pinning User to Master这种方式来实现写后读一致性

将发布信息的用户3分钟内的读操作指向master

这里暂时添加一张表来存储Pinning user，使用redis缓存其实更好，这里先不引入其他组件，减少复杂度。


# 测试结果
```sh
// 先添加人员，此测试中人员编号需要连续
# ./main -u 1000
INFO[0000] register table success
INFO[0000] command: rc1Cmd:0, rc2Cmd:0, userCmd:1000

# ./main -c2 100
INFO[0000] register table success
INFO[0000] command: rc1Cmd:0, rc2Cmd:100, userCmd:0
ERRO[0000] rcCheck2, fail, modId:487, modStr:bb6fe76a-6642-11ed-9ace-be2e1b32fdf3, readStr:
ERRO[0000] rcCheck2, fail, modId:780, modStr:bb77e11b-6642-11ed-9ace-be2e1b32fdf3, readStr:
ERRO[0000] rcCheck2, fail, modId:373, modStr:bb872e31-6642-11ed-9ace-be2e1b32fdf3, readStr:
ERRO[0000] rcCheck2, fail, modId:352, modStr:bb87e4f9-6642-11ed-9ace-be2e1b32fdf3, readStr:
INFO[0000] rc2 check done

# ./main -c1 100
INFO[0000] register table success
INFO[0000] command: rc1Cmd:100, rc2Cmd:0, userCmd:0
INFO[0000] rc1 check done

# ./main -c1 1000
INFO[0000] register table success
INFO[0000] command: rc1Cmd:1000, rc2Cmd:0, userCmd:0
INFO[0005] rc1 check done
```