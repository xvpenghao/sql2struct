# sql2struct

* sql 转换为 xorm的model
* 通过解析 xxx.toml配置，读取 创建表信息然后 生成 对应 的xorm model 和对应的实现方法



# Quick start

安装 `sql2s `工具 ，或者 下载源码编译安装



下载 sql2工具

```
go get -u github.com/xvpneghao/sql2struct/cmd/sql2s
```

写配置文件

```toml
dsn="root:123456@tcp(localhost:3306)/db_test?charset=utf8"
dstFile="example/score.go"
structName="TotalScore"
pkgName="mscore"
tableName="t_score_total"
```

运行 sql2并 指定配置文件路径，即可生成 xxx.go文件

```shell
./sql2s -src=xxxx.toml
```


