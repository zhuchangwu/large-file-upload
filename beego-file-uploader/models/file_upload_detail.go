package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type FileUploadDetail struct {
	Id              int       `json:"id" orm:"column(id) pk;auto"`
	Username        string    `json:"username" orm:"column(username) size(64)"`
	FileName        string    `json:"fileName" orm:"column(file_name) size(64)"`
	Md5             string    `json:"md5" orm:"column(md5) size(255)"`
	IsUploaded      int       `json:"isUploaded" orm:"column(is_uploaded)"`
	TotalChunks     int       `json:"totalChunks" orm:"column(total_chunks)"`
	HasBeenUploaded string    `json:"hasbeenUploaded" orm:"column(hasbeen_uploaded) size(1024)"`
	Url             string    `json:"url" orm:"column(url) size(255)" `
	CreateTime      time.Time `json:"createTime" orm:"column(create_time) type(datetime);auto_now_add"`
	UpdateTime      time.Time `json:"updateTime" orm:"column(update_time) type(datetime);"`
}

func (f *FileUploadDetail) TableName() string {
	return "file_upload_detail"
}

// 根据username 和 fileName查询文件的细节
func (f *FileUploadDetail) FindUploadDetailByFileName(username, fileName string) (detail FileUploadDetail, err error) {
	newOrmer := orm.NewOrm()
	err = newOrmer.QueryTable(&FileUploadDetail{}).Filter("username", username).Filter("fileName", fileName).One(&detail)
	return
}

// 保存一条记录
func (f *FileUploadDetail) InsertOneRecord() (id int64, err error) {
	id, err = orm.NewOrm().Insert(f)
	return
}

// 更新一列
func (f *FileUploadDetail) UpdateColumn(uploaded string) (num int64, err error) {
	POrmer := orm.NewOrm()
	num, err = POrmer.Update(f, uploaded)
	if err != nil {
		fmt.Printf("fail to update fileUploadDetail cloumn ,columnName:[%v]", uploaded)
		return
	}
	return
}

func NewFileUploadDetail() *FileUploadDetail {
	return &FileUploadDetail{}
}
