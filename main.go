package main

import (
	_Struct "FileManagement/Struct"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"net/url"

	"github.com/gin-gonic/gin"
)

// 合并切片
func mergeChunks(fileName string, totalChunks int, finalDst string) {
	out, _ := os.Create(finalDst)
	defer out.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := fmt.Sprintf("./temp/%s_%d", fileName, i)
		chunk, _ := os.Open(chunkPath)
		io.Copy(out, chunk)
		chunk.Close()
		os.Remove(chunkPath) // 删除临时切片文件
	}
}

func main() {
	// 确保上传目录和临时目录存在
	os.MkdirAll("./uploaded", 0755)
	os.MkdirAll("./temp", 0755)

	r := gin.Default()
	r.LoadHTMLGlob("./templates/*")
	// 创建一个文件服务
	var fileService _Struct.FileService
	// 初始化文件服务
	err := fileService.InitfileService("./uploaded")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 跳转到首页（文件管理）
	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"fileList": fileService.FileDocuments,
		})
	})

	// 跳转到上传页面
	r.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{
			"fileList": fileService.FileDocuments,
		})
	})

	// 文件上传(小文件)
	r.POST("/upload", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		files := form.File["f1"]

		fmt.Println(files)

		// 将上传的文件下载至指定目录中
		for _, file := range files {
			dst := fmt.Sprintf("./uploaded/%s", file.Filename)
			err := c.SaveUploadedFile(file, dst)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message": "上传文件失败" + err.Error(),
				})
				return
			}
			// 更新文件服务信息
			fileService.FileDocuments = append(fileService.FileDocuments, _Struct.FileDocument{
				Id:   len(fileService.FileDocuments) + 1,
				Name: file.Filename,
				Type: filepath.Ext(file.Filename),
			})
		}
		fmt.Println("调用普通文件上传方式")
		// 返回更新后的文件列表
		c.HTML(http.StatusOK, "upload.html", gin.H{
			"fileList": fileService.FileDocuments,
		})
	})

	// 文件切片上传（大文件）
	r.POST("/upload-chunk", func(c *gin.Context) {
		chunkNumber, _ := strconv.Atoi(c.PostForm("chunkNumber"))
		totalChunks, _ := strconv.Atoi(c.PostForm("totalChunks"))
		fileName := c.PostForm("fileName")

		// 保存切片到临时目录
		chunk, _ := c.FormFile("chunk")
		dst := fmt.Sprintf("./temp/%s_%d", fileName, chunkNumber)
		c.SaveUploadedFile(chunk, dst)

		// 检查所有切片是否都已上传完成
		if chunkNumber == totalChunks-1 {
			// 合并切片 指定切片最终保存位置
			finalDst := fmt.Sprintf("./uploaded/%s", fileName)
			mergeChunks(fileName, totalChunks, finalDst)

			// 更新文件服务信息
			fileService.FileDocuments = append(fileService.FileDocuments, _Struct.FileDocument{
				Id:   len(fileService.FileDocuments) + 1,
				Name: fileName,
				Type: filepath.Ext(fileName),
			})
		}
		fmt.Println("调用切片文件上传方式")

		// 返回更新后的文件列表
		c.JSON(http.StatusOK, gin.H{
			"message":  "切片上传成功",
			"fileList": fileService.FileDocuments,
		})
	})

	// 删除上传文件
	r.DELETE("/delete/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		fmt.Println(idStr)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "无效的ID",
			})
			return
		}
		// 获取文件信息并删除物理文件（删除文件夹中文件） 主要是获取文件名字
		file := fileService.GetFileById(id)
		if file != nil {
			// 删除物理文件
			err = os.Remove(fmt.Sprintf("./uploaded/%s", file.Name))
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "删除文件失败: " + err.Error(),
				})
				return
			}
		}

		// 根据 id获取对象 然后删除文件状态（显示）
		fileService.DeleteFileById(id)
		c.JSON(200, gin.H{
			"success": true,
			"message": "删除成功",
		})
	})

	// 跳转至下载文件页面
	r.GET("/download", func(c *gin.Context) {
		c.HTML(http.StatusOK, "download.html", gin.H{
			"fileList": fileService.FileDocuments,
		})
	})

	// 下载文件 断点下载
	r.GET("/download/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		fmt.Println(idStr)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "无效的ID",
			})
			return
		}
		// 根据ID获取文件信息
		file := fileService.GetFileById(id)
		if file == nil {
			c.JSON(404, gin.H{
				"success": false,
				"message": "文件不存在",
			})
			return
		}

		// 构建文件完整路径
		filePath := fmt.Sprintf("./uploaded/%s", file.Name)

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(404, gin.H{
				"success": false,
				"message": "文件不存在于服务器",
			})
			return
		}

		// 对文件名进行URL编码，解决中文文件名问题
		encodedFileName := url.QueryEscape(file.Name)

		// 设置Content-Disposition头，同时兼容不同浏览器
		c.Writer.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
				file.Name, encodedFileName))

		// 设置响应类型
		c.Writer.Header().Set("Content-Type", "application/octet-stream")

		// 使用http.ServeFile支持断点续传
		http.ServeFile(c.Writer, c.Request, filePath)
	})

	r.Run()
}
