package mscore

import (
	"time"
)

type TotalScore struct {
	Id         int    `zh:"" json:"id" form:"id" xorm:"'id' bigint pk autoincr"`
	BookId     int    `zh:"图书id" json:"bookId" form:"bookId"`
	TotalCount int    `zh:"实际总人数" json:"totalCount" form:"totalCount"`
	TotalScore int    `zh:"实际总评分" json:"totalScore" form:"totalScore"`
	LevelScore int    `zh:"level分数" json:"levelScore" form:"levelScore"`
	CreateTime string `zh:"创建时间" json:"createTime" form:"createTime"`
	UpdateTime string `zh:"更新时间" json:"updateTime" form:"updateTime"`
}

func (this *TotalScore) TableName() string {
	return "t_score_total"
}

func (this *TotalScore) BeforeInsert() {
	if this.CreateTime == "" {
		this.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	}
	if this.UpdateTime == "" {
		this.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	}
}

func (this *TotalScore) BeforeUpdate() {
	this.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
}
