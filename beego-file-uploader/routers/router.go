package routers

import (
	"github.com/astaxie/beego/plugins/cors"
	"oa-flow-centor/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowOrigins:     []string{"http://192.168.43.52"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type", "X-TOKEN"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	beego.Router("/file/upload", &controllers.FileUploadController{}, "*:Upload")
	beego.Router("/file/merge", &controllers.FileUploadController{}, "*:Merge")
}
