package model

type CreateTableSql struct {
	TableName  string    `json:"name"`
	ColumnList []*Column `json:"columnList"`
}

type Column struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Comment  string `json:"comment"`
}
