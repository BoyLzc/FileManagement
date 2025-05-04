package _Struct

import (
	"fmt"
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
	Size string `json:"size"`
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

		fileSize := FormatFileSize(fileInfo.Size())

		// 将文件信息添加到切片中
		fileList = append(fileList, FileDocument{
			Id:   i + 1,
			Name: fileInfo.Name(),
			Type: fileType,
			Size: fileSize,
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

// 统计当前服务的文件列表总数
func (fileService *FileService) GetFileList() int {
	return len(fileService.FileDocuments)
}

// 获取文件总大小（字节数）
func (fileService *FileService) GetTotalSize(basePath string) (int64, error) {
	var totalSize int64 = 0
	for _, file := range fileService.FileDocuments {
		// 构建文件完整路径
		filePath := filepath.Join(basePath, file.Name)
		// 获取文件信息
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			// 如果文件不存在或无法访问，跳过该文件
			continue
		}
		// 累加文件大小
		totalSize += fileInfo.Size()
	}
	return totalSize, nil
}

// 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// 获取格式化的文件总大小（如：1.2 MB）
func (fileService *FileService) GetFormattedTotalSize(basePath string) (string, error) {
	totalSize, err := fileService.GetTotalSize(basePath)
	if err != nil {
		return "", err
	}
	return FormatFileSize(totalSize), nil
}
