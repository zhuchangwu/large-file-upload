package controllers

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"oa-flow-centor/models"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileUploadController struct {
	BaseController
}

type FileUploadInfo struct {
	Id               int    `json:"id"`
	ChunkNumber      int    `json:"chunkNumber"`      // 当前是第几片
	ChunkSize        int    `json:"chunkSize"`        // 每片的大小
	CurrentChunkSize int    `json:"currentChunkSize"` // 当前分片的大小
	TotalSize        int    `json:"totalSize"`        // 文件的总大小
	TotalChunks      int    `json:"totalChunks"`      // 总分片数
	FileName         string `json:"fileName"`
	Identifier       string `json:"identifier"` // fileMd5值
	HasBeenUploaded  string `json:"hasBeenUploaded"`
}

// todo 最好再加个状态码
// 响应信息
type FileUploadResponseInfo struct {
	IsUploaded      int    `json:"isUploaded"`      // 是否已经完成上传了 0:否  1:是 - 秒传
	HasBeenUploaded string `json:"hasBeenUploaded"` // 曾经上传过的分片chunkNumber - 断点续传
	Merge           int    `json:"merge"`           // 是否可以合并了   0：否  1:是
	Status          int    `json:"status"`          // 是否可以合并了   0：成功  1:失败
	Msg             string `json:"msg"`             // 其他信息
	URL             string `json:"url"`             // 后端保存的url
}

// todo 把下面的0、1这种数字抽取到common中，改成常量的编码形式
// todo 最好将 在解析保存chunk的过程所报的错误返回给前端
// todo 将updateColumn的动作 紧挨着 saveChunk成功后
/**
 *  约定：分片的命名规则    /upload/username/fileMd5_chunkNumber.Ext
 *  约定：merge后的文件名  /upload/username/fileMd5.Ext
 */
func (c *FileUploadController) Upload() {

	// 返回给前端的对象
	resultInfo := FileUploadResponseInfo{}
	// 接收，处理数据
	filename := c.GetString("filename")
	chunkNumber, _ := c.GetInt("chunkNumber")
	currentChunkSize, _ := c.GetInt("currentChunkSize")
	totalChunks, _ := c.GetInt("totalChunks")
	fileMd5 := c.GetString("identifier")
	username := "changwu"
	tempFileName := fileMd5 + "_" + strconv.Itoa(chunkNumber) + filepath.Ext(filename) // 当前的分片名
	targetFileName := fileMd5 + filepath.Ext(filename)                                 // 最终的文件名
	// 分片校验
	method := c.Ctx.Request.Method

	// 上传前的预检请求
	if method == "GET" {
		detail := models.FileUploadDetail{}
		detail, err := detail.FindUploadDetailByFileName(username, targetFileName)

		// 查询不到记录，说明未曾这是第一次上传
		if err != nil && err.Error() == "<QuerySeter> no row found" {
			// 更新数据库的hasBeenUploaded
			detail := &models.FileUploadDetail{
				Username:        username,
				FileName:        targetFileName,
				Md5:             fileMd5,
				TotalChunks:     totalChunks,
				IsUploaded:      0,
				HasBeenUploaded: "",
				CreateTime:      time.Now(),
				UpdateTime:      time.Now(),
			}
			id, err := detail.InsertOneRecord()
			if err != nil || id < 0 {
				fmt.Printf("fail to insert upload detail record fineName:[%v] md5:[%v] err:[%v]", filename, fileMd5, err)
			}
			// 如果单个chunk就可以完成文件的上传，直接告诉前端可以合并了
			// 前端单线程顺序发送chunk，故下面的条件衡成立
			if chunkNumber == totalChunks {
				resultInfo.IsUploaded = 0
				resultInfo.Merge = 1
				c.Data["json"] = resultInfo
				c.ServeJSON()
				return
			}

			resultInfo.IsUploaded = 0
			resultInfo.Merge = 0
			resultInfo.HasBeenUploaded = "" // 告诉前端第一片chunk上传成功了
			c.Data["json"] = resultInfo
			c.ServeJSON()
			return
		}
		// 校验一下，当前文件是否曾经上传过，并且完整的上传完成了
		if detail.IsUploaded == 1 { // 已经上传过了
			resultInfo.IsUploaded = 1
			resultInfo.Merge = 0
			// todo 改变状态码，非200
			c.Data["json"] = resultInfo
			c.ServeJSON()
			return
		}

		// 处理断点续传chunk
		resultInfo.IsUploaded = 0
		resultInfo.Merge = 0
		resultInfo.HasBeenUploaded = detail.HasBeenUploaded // 将曾经上传过的记录发送给前端
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}

	// post请求
	detail, err := models.NewFileUploadDetail().FindUploadDetailByFileName(username, targetFileName)
	if err != nil && err.Error() == "<QuerySeter> no row found" {
		fmt.Printf("post: fail to findUploadDetailByFileName username:[%v] targetFileName:[%v]", username, targetFileName)
		return
	}

	// 判断是否已经上传成了～
	if detail.IsUploaded == 1 {

		// todo 判断是否曾经merge过

		fmt.Printf("post: username: [%v] has been uploaded the file fileName:[%v] md5:[%v]", username, targetFileName, fileMd5)
		resultInfo.IsUploaded = 1
		resultInfo.Merge = 0
		// todo 改变状态码，非200
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}

	// todo 当前currentSize为空， 需要特殊处理一下

	// 保存当前chunk
	err = SaveChunkToLocalFromMutipartForm(c, tempFileName, username, currentChunkSize)
	if err != nil {
		fmt.Printf("post fail to save chunk fileName:[%v] md5:[%v] chunkNumber:[%v]", filename, fileMd5, chunkNumber)
		// 告诉前端重传
		c.Ctx.ResponseWriter.WriteHeader(500)
		resultInfo.IsUploaded = 1
		resultInfo.Merge = 0
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}
	// 多协程并发修改单行数据会产生覆盖，但是前端会单线程访问，故下面的操作安全
	if detail.TotalChunks == chunkNumber {
		detail.HasBeenUploaded = detail.HasBeenUploaded + strconv.Itoa(chunkNumber)
	} else {
		detail.HasBeenUploaded = detail.HasBeenUploaded + strconv.Itoa(chunkNumber) + ":"
	}
	num, err := detail.UpdateColumn("has_been_uploaded")
	if err != nil || num < 0 {
		fmt.Printf("fail to updateColumn err:[%v]", err)
		return
	}

	// 如果前端是并发上传的化，这个顺序不能保证～，所以下面的条件应该是比对 totalChunks == chunkNumber不能保证全部正确
	// 目前前端会采用单条线程顺序发送chunk，故下面的条件衡成立
	// 服务端判断，当前chunkNumber == totalChunks时, 先保存当前chunk 再向前端发送特殊响应，前端接收到后会发送merge请求
	if chunkNumber == totalChunks {
		// 更新数据库 标记文件全部上传完成
		detail.IsUploaded = 1
		num, err = detail.UpdateColumn("is_uploaded")
		if err != nil || num < 0 {
			fmt.Printf("fail to updateColumn err:[%v]", err)
			return
		}
		resultInfo.IsUploaded = 0
		// 告诉前端可以merge了
		resultInfo.Merge = 1
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}

	resultInfo.IsUploaded = 0
	resultInfo.Merge = 0
	resultInfo.Msg = "ok"
	c.Data["json"] = resultInfo
	c.ServeJSON()
	return
}

/**
 *   desc：
 *       将chunk中的数据暂时存在本地
 *
 *	 params：
 *	 	 tempFileName: 当前分片使用的文件名
 *		 currentChunkSize: 当前分片的大小
 */
func SaveChunkToLocalFromMutipartForm(c *FileUploadController, tempFileName, username string, currentChunkSize int) (err error) {
	// 创建文件夹
	path, err := os.Getwd()
	folderPath := path + "/upload/" + username
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		fmt.Println("创建文件夹")
		// 必须分成两步
		// 先创建文件夹
		os.MkdirAll(folderPath, 0777)

		// 再修改权限
		os.Chmod(folderPath, 0777)

	}

	// 保存本次上传的分片 /upload/username/fileMd5_chunkNumber.suffix
	// todo 为什么第一次是空的 (因为第一次是get)
	if c.Ctx.Request.MultipartForm == nil {
		return
	}
	fileHeader := c.Ctx.Request.MultipartForm.File["file"][0]
	if fileHeader == nil {
		err = errors.New("fileHeader 为空")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	// 在本地创建文件，如果没有就创建，如果有就打开
	myFile, err := os.Create(folderPath + "/" + tempFileName)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}

	// 循环读取客户端发送过来的文件
	buf := make([]byte, currentChunkSize)
	num, err := file.Read(buf)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	fmt.Printf("本轮读取了 num=[%v] byte", num)

	// 保存文件
	num, err = myFile.Write(buf)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	fmt.Printf("本次保存分片Size为 num == [%v] ", num)
	// 关闭文件
	myFile.Close()
	file.Close()
	return
}

/**
 * desc：merge files
 *
 * params：
 *		fileName
 *	    fileMd5
 *
 * return：
 *		msg: mergeDone:成功合并  /mergeErr：合并过程中出错 /fileMd5Err：fileMd5比对失败了
 *      url: sourceUrl
 */
func (c *FileUploadController) Merge() {
	// 返回给前端的对象
	resultInfo := FileUploadResponseInfo{}

	fileMd5 := c.GetString("identifier")
	fileName := c.GetString("fileName")
	username := "changwu"
	targetFileName := fileMd5 + filepath.Ext(fileName)
	// 根据username + md5 查询数据库
	detail, err := models.NewFileUploadDetail().FindUploadDetailByFileName(username, targetFileName)
	if err != nil && err.Error() == "<QuerySeter> no row found" {
		fmt.Printf("merge: fail to findUploadDetailByFileName username:[%v] targetFileName:[%v]", username, targetFileName)
		return
	}

	// 校验detail的isUploaded字段是否为1
	if detail.IsUploaded != 1 {
		str := fmt.Sprintf("fail to merge chunks because filename:[%v] isUploaded:[%v]", targetFileName, detail.IsUploaded)
		resultInfo.Msg = str
		resultInfo.Status = 1
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}
	// 校验detail的totalChunk是否和 hasBeanUploaded中按照 : 分割后得到的结果数相同
	arr := strings.Split(detail.HasBeenUploaded, ":")
	if detail.TotalChunks != len(arr) {
		str := fmt.Sprintf("fail to merge chunks because some fragments that have not been uploaded filename:[%v] totalChunks:[%v] hasBeenUploaded:[%v]", targetFileName, detail.TotalChunks, detail.HasBeenUploaded)
		resultInfo.Msg = str
		resultInfo.Status = 1
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}

	// 开始merge

	dir, _ := os.Getwd()
	baseUrl := dir + "/upload/" + username + "/"
	targetFileName = baseUrl + fileMd5 + filepath.Ext(fileName)

	// todo 考虑这里要不要查看下数据库中isUploaded的信息

	// 创建临时文件
	f, err := os.OpenFile(targetFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	var totalSize int64
	writer := bufio.NewWriter(f)
	// 合并文件  /upload/username/fileMd5/chunks
	// 读出指定目录下的所有文件
	for i := 1; i <= detail.TotalChunks; i++ {
		currentChunkFile := baseUrl + fileMd5 + "_" + strconv.Itoa(i) + filepath.Ext(fileName) // 当前的分片名
		bytes, err := ioutil.ReadFile(currentChunkFile)
		if err != nil {
			fmt.Printf("error : %v", err)
			return
		}
		num, err := writer.Write(bytes)
		if err != nil {
			fmt.Printf("error : %v", err)
			return
		}
		totalSize += int64(num)
		fmt.Printf("本次读取了 %v", num)
		// todo 暂时选择移除, 等所有的工作都做完了，可以考虑这样：第一次merge失败了我们也不会删除，提示用户重新上传就好了，并在数据库中记录曾经merge失败过
		// todo 如果第二次merge又失败了，删除暂存的分片，删除用户的上传记录
		err = os.Remove(currentChunkFile)
		if err != nil {
			fmt.Printf("error : %v", err)
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	// 在重新打开文件之间关闭
	f.Close()

	// 重新打开文件，目的是获取到最新的md5
	f, err = os.OpenFile(targetFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	defer f.Close()

	// 计算合并后文件的fileMd5
	md5 := md5.New()
	writtenNum, err := io.Copy(md5, f)
	if err != nil {
		fmt.Printf("Fail file copy into : %v", err)
		return
	}
	fmt.Println("file copy into md5 bytes:[%v]", writtenNum)
	md5Val := hex.EncodeToString(md5.Sum(nil))

	// 校验md5
	if md5Val != detail.Md5 {
		// todo 删除文件及数据库中的记录
		// md5不同
		resultInfo.Msg = "fileMd5Err"
		resultInfo.Status = 1
		c.Data["json"] = resultInfo
		c.ServeJSON()
		return
	}
	fmt.Println("file md5 is lawful ，will return fileUrl to fontEnd")
	// todo 将url更新到数据库
	// 合并成功，给前端响应
	resultInfo.Msg = "mergeDone"
	resultInfo.URL = "this is fileUrl"
	//resultInfo.URL = targetFileName
	c.Data["json"] = resultInfo
	c.ServeJSON()
	return
}
