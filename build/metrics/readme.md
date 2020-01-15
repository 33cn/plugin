
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

* 数据的可视化展示可以通过Grafana工具进行展示，只需要访问以下链接：
  http://build_grafana_1:3000，更多的操作可以参考以下链接中的使用Grafana工具展示部分

