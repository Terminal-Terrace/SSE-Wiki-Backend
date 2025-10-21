package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	dbpkg "terminal-terrace/sse-wiki/internal/database"
	filemodel "terminal-terrace/sse-wiki/internal/model/file"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Init(c *gin.Context) {
	var req InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 秒传检查
	var existing filemodel.File
	if err := h.db.Where("file_hash = ?", req.FileHash).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, InitUploadResponse{Exists: true, FileID: existing.ID, FileURL: fmt.Sprintf("/api/v1/files/%d", existing.ID)})
		return
	}

	// 2. 创建会话
	uploadID := uuid.NewString()
	session := UploadSession{
		UploadID:       uploadID,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		FileHash:       req.FileHash,
		TotalChunks:    req.TotalChunks,
		UploadedChunks: make([]bool, req.TotalChunks),
		MimeType:       req.MimeType,
	}

	sessionJSON, _ := json.Marshal(session)
	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()
	if err := dbpkg.RedisDB.Set(ctx, "upload:"+uploadID, sessionJSON, 2*time.Hour).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存会话失败"})
		return
	}

	// 3. 创建临时目录
	if err := os.MkdirAll(filepath.Join("temp", uploadID), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建临时目录失败"})
		return
	}

	c.JSON(http.StatusOK, InitUploadResponse{Exists: false, UploadID: uploadID})
}

func (h *Handler) Chunk(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	chunkIndexStr := c.PostForm("chunkIndex")
	if uploadID == "" || chunkIndexStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少uploadId或chunkIndex"})
		return
	}
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的chunkIndex"})
		return
	}

	// 获取文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件分块"})
		return
	}

	// 读取会话
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()
	sessionJSON, err := dbpkg.RedisDB.Get(ctx, "upload:"+uploadID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "上传会话不存在或已过期"})
		return
	}
	var session UploadSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "会话解析失败"})
		return
	}
	if chunkIndex >= session.TotalChunks {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunkIndex超出范围"})
		return
	}

	// 保存分块
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取分块失败"})
		return
	}
	defer src.Close()

	chunkPath := filepath.Join("temp", uploadID, fmt.Sprintf("chunk_%d", chunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存分块失败"})
		return
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入分块失败"})
		return
	}
	dst.Close()

	// 更新会话
	session.UploadedChunks[chunkIndex] = true
	sessionJSONBytes, _ := json.Marshal(session)
	if err := dbpkg.RedisDB.Set(ctx, "upload:"+uploadID, sessionJSONBytes, 2*time.Hour).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (h *Handler) Complete(c *gin.Context) {
	var req CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	sessionJSON, err := dbpkg.RedisDB.Get(ctx, "upload:"+req.UploadID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "上传会话不存在或已过期"})
		return
	}

	var session UploadSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "会话解析失败"})
		return
	}

	// 校验完整性
	for i, ok := range session.UploadedChunks {
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("分块 %d 缺失", i)})
			return
		}
	}

	// 合并分块
	ext := filepath.Ext(session.FileName)
	finalPath := filepath.Join("uploads", session.FileHash+ext)
	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建uploads目录失败"})
		return
	}

	finalFile, err := os.Create(finalPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目标文件失败"})
		return
	}
	defer finalFile.Close()

	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join("temp", req.UploadID, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打开分块失败"})
			return
		}
		if _, err := io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "合并分块失败"})
			return
		}
		chunkFile.Close()
	}

	// 验证大小
	stat, err := os.Stat(finalPath)
	if err != nil || stat.Size() != session.FileSize {
		_ = os.Remove(finalPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件大小不匹配"})
		return
	}

	// 插入数据库（去重，防并发）
	fileRec := filemodel.File{
		FileName:   session.FileName,
		FileHash:   session.FileHash,
		FilePath:   finalPath,
		FileSize:   session.FileSize,
		MimeType:   session.MimeType,
		Category:   inferCategory(session.MimeType),
		Extension:  ext,
		UploadedBy: 0, // TODO: 从JWT解析用户ID
	}

	// 先查，避免重复
	var existing filemodel.File
	if err := h.db.Where("file_hash = ?", session.FileHash).First(&existing).Error; err == nil {
		// 已存在则复用
		_ = os.Remove(finalPath)
		// 清理临时
		_ = os.RemoveAll(filepath.Join("temp", req.UploadID))
		_ = dbpkg.RedisDB.Del(ctx, "upload:"+req.UploadID).Err()
		c.JSON(http.StatusOK, gin.H{"fileId": existing.ID, "fileName": existing.FileName, "fileUrl": fmt.Sprintf("/api/v1/files/%d", existing.ID), "category": existing.Category})
		return
	}

	if err := h.db.Create(&fileRec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入数据库失败"})
		return
	}

	// 清理
	_ = os.RemoveAll(filepath.Join("temp", req.UploadID))
	_ = dbpkg.RedisDB.Del(ctx, "upload:"+req.UploadID).Err()

	c.JSON(http.StatusOK, gin.H{"fileId": fileRec.ID, "fileName": fileRec.FileName, "fileUrl": fmt.Sprintf("/api/v1/files/%d", fileRec.ID), "category": fileRec.Category})
}

func inferCategory(mime string) string {
	switch {
	case len(mime) >= 6 && mime[:6] == "image/":
		return "image"
	case len(mime) >= 6 && mime[:6] == "video/":
		return "video"
	case len(mime) >= 6 && mime[:6] == "audio/":
		return "audio"
	default:
		return "document"
	}
}


