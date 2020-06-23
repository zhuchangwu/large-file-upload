package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init(){
	// 注册驱动～
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}

	// 注册DB
	err = orm.RegisterDataBase("default", "mysql", "root:root@tcp(localhost:3306)/oa?charset=utf8")
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}

	// 注册modal
	orm.RegisterModel(
		new(FileUploadDetail),
	)

	//// 自动建表
	//err = orm.RunSyncdb("default", false, true)
	//if err != nil {
	//	fmt.Printf("error : %v", err)
	//	return
	//}
	//orm.RunCommand()
}
