package file

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// GetFile 获取文件（在线预览）
// GET /api/files/:id
func (h *Handler) GetFile(c *gin.Context) {
	// 1. 解析文件 ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件 ID"})
		return
	}

	// 2. 查询数据库获取文件元数据
	var fileRecord struct {
		ID       uint   `gorm:"column:id"`
		FileName string `gorm:"column:file_name"`
		FilePath string `gorm:"column:file_path"`
		MimeType string `gorm:"column:mime_type"`
		FileSize int64  `gorm:"column:file_size"`
	}

	err = h.db.Table("files").
		Select("id, file_name, file_path, mime_type, file_size").
		Where("id = ?", fileID).
		First(&fileRecord).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	// 3. 打开文件系统中的实际文件
	file, err := os.Open(fileRecord.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}
	defer file.Close()

	// 4. 设置响应头（在线预览模式）
	c.Header("Content-Type", fileRecord.MimeType)
	c.Header("Content-Disposition", `inline; filename="`+fileRecord.FileName+`"`)
	c.Header("Content-Length", strconv.FormatInt(fileRecord.FileSize, 10))

	// 缓存策略（文件不会改变，可以长期缓存）
	c.Header("Cache-Control", "public, max-age=31536000") // 1年
	c.Header("ETag", strconv.FormatUint(uint64(fileRecord.ID), 10))

	// 5. 流式返回文件内容（不会一次性读入内存）
	c.Status(http.StatusOK)
	io.Copy(c.Writer, file)
}

// DownloadFile 下载文件（强制下载）
// GET /api/files/:id/download
func (h *Handler) DownloadFile(c *gin.Context) {
	// 1. 解析文件 ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件 ID"})
		return
	}

	// 2. 查询数据库获取文件元数据
	var fileRecord struct {
		ID       uint   `gorm:"column:id"`
		FileName string `gorm:"column:file_name"`
		FilePath string `gorm:"column:file_path"`
		FileSize int64  `gorm:"column:file_size"`
	}

	err = h.db.Table("files").
		Select("id, file_name, file_path, file_size").
		Where("id = ?", fileID).
		First(&fileRecord).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	// 3. 打开文件系统中的实际文件
	file, err := os.Open(fileRecord.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}
	defer file.Close()

	// 4. 设置响应头（强制下载模式）
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", `attachment; filename="`+fileRecord.FileName+`"`)
	c.Header("Content-Length", strconv.FormatInt(fileRecord.FileSize, 10))

	// 5. 增加下载计数（异步执行，不影响响应速度）
	go func() {
		h.db.Table("files").
			Where("id = ?", fileID).
			UpdateColumn("download_count", gorm.Expr("download_count + 1"))
	}()

	// 6. 流式返回文件内容
	c.Status(http.StatusOK)
	io.Copy(c.Writer, file)
}

// GetFileInfo 获取文件信息（不返回文件内容）
// GET /api/files/:id/info
func (h *Handler) GetFileInfo(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件 ID"})
		return
	}

	var fileRecord struct {
		ID            uint   `json:"id"`
		FileName      string `json:"fileName"`
		FileSize      int64  `json:"fileSize"`
		MimeType      string `json:"mimeType"`
		Category      string `json:"category"`
		FileUrl       string `json:"fileUrl"`
		DownloadCount uint   `json:"downloadCount"`
		CreatedAt     string `json:"createdAt"`
	}

	err = h.db.Table("files").
		Select("id, file_name, file_size, mime_type, category, download_count, created_at").
		Where("id = ?", fileID).
		First(&fileRecord).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	// 构造文件 URL
	fileRecord.FileUrl = "/api/files/" + strconv.FormatUint(uint64(fileRecord.ID), 10)

	c.JSON(http.StatusOK, fileRecord)
}
