# 源码分析

## 目录结构及功能

```shell
audit
build
ctl
fakedb
monitor
planner
proxy # neodb 的核心逻辑目录
    query.go # 其中的 ComQuery 函数是处理SQL的核心逻辑
router
xcontext
backend
config
executor
fuzz
optimizer
plugins
neodb
    neodb.go # neodb 组件的入口
syncer
xbase
vendor
```
