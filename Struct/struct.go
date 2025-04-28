package _Struct

import (
	"os"
	"path/filepath"
)

// 文件服务
type FileService struct {
	FileDocuments []FileDocument `json:"fileDocuments"`
}

// 创建文件信息结构体
type FileDocument struct {
	Id   int    `json:"Id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// 添加分页相关结构体
type PageInfo struct {
	FileList    []FileDocument `json:"fileList"`
	CurrentPage int            `json:"currentPage"`
	PageSize    int            `json:"pageSize"`
	Total       int            `json:"total"`
	TotalPages  int            `json:"totalPages"`
}

// 初始化文件服务，读取文件中信息
func (fileService *FileService) InitfileService(path string) error {
	// 读取 uploaded 目录
	files, err := os.ReadDir(path)

	// 创建文件信息切片
	var fileList []FileDocument

	// 遍历文件夹中的所有文件
	for i, file := range files {
		// 获取文件信息
		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		// 获取文件扩展名
		fileType := filepath.Ext(fileInfo.Name())
		if fileType == "" {
			fileType = "unknown"
		}

		// 将文件信息添加到切片中
		fileList = append(fileList, FileDocument{
			Id:   i + 1,
			Name: fileInfo.Name(),
			Type: fileType,
		})
	}
	fileService.FileDocuments = fileList
	return err
}

// 根据文件id删除上传文件
func (fileService *FileService) DeleteFileById(id int) {
	// 根据 id获取对象 然后改变book的状态
	for i, file := range fileService.FileDocuments {
		if file.Id == id {
			fileService.FileDocuments = append(fileService.FileDocuments[:i], fileService.FileDocuments[i+1:]...)
			break
		}
	}
}

// 根据文件id获取文件信息
func (fileService *FileService) GetFileById(id int) *FileDocument {
	// 根据 id获取对象
	for _, file := range fileService.FileDocuments {
		if file.Id == id {
			return &file
		}
	}
	return nil
}
