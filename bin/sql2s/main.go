package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pelletier/go-toml"
	"github.com/xvpenghao/sql2struct/model"
	"github.com/xvpenghao/sql2struct/templates"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"xorm.io/xorm"
)

func main() {
	srcPath := flag.String("src", "", " config file path -src=xxx.toml")
	flag.Parse()
	if *srcPath == "" {
		log.Fatal("input xxx.toml file path")
		return
	}

	data, err := ioutil.ReadFile(*srcPath)
	if err != nil {
		log.Fatal(err)
	}

	cfg := new(Config)
	if err := toml.Unmarshal(data, cfg); err != nil {
		log.Fatal(err)
	}
	g(cfg)
}

type Config struct {
	DSN        string `toml:"dsn"`
	DstFile    string `toml:"dstFile"`
	StructName string `toml:"structName"`
	PkgName    string `toml:"pkgName"`
	TableName  string `toml:"tableName"`
}

func g(cfg *Config) {
	engine, err := xorm.NewEngine("mysql", cfg.DSN)
	if err != nil {
		log.Fatal(err)
		return
	}
	queryRes, err := engine.QueryString(fmt.Sprintf("show create table %s", cfg.TableName))
	if err != nil {
		log.Fatal(err)
		return
	}

	createSql := queryRes[0]["Create Table"]
	lastIndex := strings.LastIndex(createSql, "',")
	index := strings.Index(createSql, "(")
	content := strings.Split(createSql[index+1:lastIndex+1], ",")
	ctsql := &model.CreateTableSql{
		TableName: cfg.TableName,
	}
	var columnList []*model.Column
	for _, s := range content {
		s = strings.TrimSpace(s)
		row := strings.Split(s, " ")
		columnList = append(columnList, &model.Column{
			Name:     snakeCaseToCamel(strings.Trim(row[0], "`")),
			DataType: sqlType2GoType(getColType(row[1])),
			Comment:  getComment(row),
		})
	}

	ctsql.ColumnList = columnList
	b := bytes.NewBufferString(templates.GenerateModelFile(ctsql, cfg.StructName, cfg.PkgName))
	// 格式化
	formatRes, _ := format.Source(b.Bytes())
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

func getComment(row []string) string {
	// 如果里面没有 comment则直接返回
	if !strings.Contains(strings.Join(row, ","), "COMMENT") {
		return ""
	}
	return strings.Trim(row[len(row)-1], "'")
}

func sqlType2GoType(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "varchar", "char", "datetime":
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
