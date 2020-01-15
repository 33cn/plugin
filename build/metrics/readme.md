
# Metrics 功能

* 通过make metrics来测试plugin的metrics数据收集功能
* 在容器build_influxdb_1中运行以下语句，查看结果：
```
1.influx 进入influxdb交互界面
```
```
2.use chain33metrics
```
```
3.show field keys
```
```
4.使用select进行查询，如select * from mesurment
```

