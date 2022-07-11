package main

import (
	"bytes"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
	"github.com/xvpenghao/sql2struct/model"
	"github.com/xvpenghao/sql2struct/templates"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"xorm.io/xorm"
)

var (
	srcPath           string
	confArgDsn        string
	confArgTableName  string
	confArgDstFile    string
	confArgPkgName    string
	confArgStructName string
)

func main() {
	app := cli.NewApp()
	app.Description = `
example:
[1] [dsn,tableName,dstFile,pkgName,structName] use way 
./sql2model --dsn 'uname:pwd@tcp(host:3306)/db?charset=utf8' \
--tableName xxx \
--dstFile xx.go \
--pkgName xxxx \
--structName xxx 

[2] src use way
./sql2model --src=xxx.toml 
toml file content:
dsn="root:123456@tcp(localhost:3306)/db_test?charset=utf8"
dstFile="./hello.go"
structName="Hello"
pkgName="hello"
tableName="t_hello"
`
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "src",
			Usage:       "config file xxx.toml path",
			Value:       "",
			Destination: &srcPath,
		},
		cli.StringFlag{
			Name:        "dsn",
			Usage:       "[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]",
			Value:       "",
			Destination: &confArgDsn,
		},
		cli.StringFlag{
			Name:        "tableName",
			Usage:       "connect db table name",
			Value:       "",
			Destination: &confArgTableName,
		},
		cli.StringFlag{
			Name:        "dstFile",
			Usage:       "generate dst go file path",
			Value:       "./hello.go",
			Destination: &confArgDstFile,
		},
		cli.StringFlag{
			Name:        "pkgName",
			Usage:       "generate dst go file package name",
			Value:       "main",
			Destination: &confArgPkgName,
		},
		cli.StringFlag{
			Name:        "structName",
			Usage:       "generate dst go file struct name",
			Value:       "Hello",
			Destination: &confArgStructName,
		},
	}
	app.Action = action
	app.Run(os.Args)
}

func action(c *cli.Context) error {
	if srcPath != "" {
		return exeGBySrcPath()
	}
	cfg := &Config{
		DSN:        confArgDsn,
		DstFile:    confArgDstFile,
		StructName: confArgStructName,
		PkgName:    confArgPkgName,
		TableName:  confArgTableName,
	}
	g(cfg)
	return nil
}

func exeGBySrcPath() error {
	cfg := new(Config)
	data, err := ioutil.ReadFile(srcPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if err = toml.Unmarshal(data, cfg); err != nil {
		fmt.Println(err)
		return err
	}
	g(cfg)
	return err
}

type Config struct {
	DSN        string `toml:"dsn"`
	DstFile    string `toml:"dstFile"`
	StructName string `toml:"structName"`
	PkgName    string `toml:"pkgName"`
	TableName  string `toml:"tableName"`
}

func setDft(cfg *Config) {
	// 处理默认值
	if cfg.PkgName == "" {
		cfg.PkgName = "main"
	}
	if cfg.StructName == "" {
		cfg.StructName = "Hello"
	}
	if cfg.DstFile == "" {
		cfg.DstFile = "./hello.go"
	}
}

func g(cfg *Config) {
	setDft(cfg)
	d, _ := toml.Marshal(cfg)
	fmt.Println(string(d))
	engine, err := xorm.NewEngine("mysql", cfg.DSN)
	if err != nil {
		fmt.Println(err)
		return
	}
	queryRes, err := engine.QueryString(fmt.Sprintf("show create table %s", cfg.TableName))
	if err != nil {
		fmt.Println(err)
		return
	}

	createSql := queryRes[0]["Create Table"]
	lastIndex := strings.LastIndex(createSql, "',")
	index := strings.Index(createSql, "(")
	content := strings.Split(createSql[index+1:lastIndex+1], ",\n")
	ctsql := &model.CreateTableSql{
		TableName: cfg.TableName,
	}
	var columnList []*model.Column
	for _, s := range content {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		row := strings.Split(s, " ")
		columnList = append(columnList, &model.Column{
			Name:     snakeCaseToCamel(strings.Trim(row[0], "`")),
			DataType: sqlType2GoType(getColType(row[1])),
			Comment:  getComment(s),
		})
	}

	ctsql.ColumnList = columnList
	b := bytes.NewBufferString(templates.GenerateModelFile(ctsql, cfg.StructName, cfg.PkgName))
	// 格式化
	formatRes, _ := format.Source(b.Bytes())
	dir := cfg.DstFile[:strings.LastIndex(cfg.DstFile, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) { // 目录不存在，则创建目录
		os.MkdirAll(dir, os.ModePerm)
	}
	err = ioutil.WriteFile(cfg.DstFile, formatRes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func getColType(colType string) string {
	if strings.Contains(colType, "(") {
		// int(11) -> int
		index := strings.Index(colType, "(")
		return colType[:index]
	}
	return colType
}

func getComment(row string) string {
	// 如果里面没有 comment则直接返回
	if !strings.Contains(row, "COMMENT") {
		return ""
	}
	res := strings.Split(row, "COMMENT")
	if len(res) < 1 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(res[1]), "'")
}

func sqlType2GoType(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "varchar", "char", "datetime", "text":
		return "string"
	case "int", "tinyint":
		return "int"
	case "bigint":
		return "int64"
	default:
		return ""
	}
}

// 可以加一个 特殊字符的校验
func snakeCaseToCamel(str string) string {
	if str == "" {
		return ""
	}
	buf := strings.Builder{}
	for i := 0; i < len(str); i++ {
		if str[i] == '_' && i+1 < len(str) {
			buf.WriteByte(str[i+1] - 32)
			i++
			continue
		}
		buf.WriteByte(str[i])
	}
	return buf.String()
}
