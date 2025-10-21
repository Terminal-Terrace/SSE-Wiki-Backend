// Package file 文件相关模型
package file

import (
	"time"
)

// File 文件信息表（存储文件元数据，不存储文件内容）
type File struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	FileName string `gorm:"type:varchar(255);not null" json:"fileName"`
	// SHA256 哈希值（64位十六进制字符串），用于秒传和去重
	FileHash string `gorm:"type:varchar(64);uniqueIndex;not null" json:"fileHash"`
	// 文件在服务器上的存储路径（相对路径，如 "uploads/a1b2c3...pdf"）
	FilePath  string `gorm:"type:varchar(500);not null" json:"filePath"`
	FileSize  int64  `gorm:"not null" json:"fileSize"`
	MimeType  string `gorm:"type:varchar(100);not null" json:"mimeType"`
	Category  string `gorm:"type:varchar(50);not null;index" json:"category"` // image/video/audio/document/archive/code/other
	Extension string `gorm:"type:varchar(20)" json:"extension"`
	// 上传者 ID
	UploadedBy uint `gorm:"not null;index" json:"uploadedBy"`
	// 下载次数统计
	DownloadCount uint      `gorm:"default:0" json:"downloadCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

// ArticleVersionFile 文章版本与文件的关联表
// 一个文章版本可以包含多个文件（内嵌图片、附件等）
// 一个文件可以被多个文章版本引用（节省存储空间）
type ArticleVersionFile struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	VersionID uint   `gorm:"not null;index:idx_version_file" json:"versionId"`
	FileID    uint   `gorm:"not null;index:idx_version_file" json:"fileId"`
	// 文件类型：inline（内嵌）、attachment（附件）、cover（封面）
	FileType string `gorm:"type:varchar(20);not null" json:"fileType"`
	// 在文章中的位置顺序（用于排序）
	Position  int       `gorm:"default:0" json:"position"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName 指定表名
func (ArticleVersionFile) TableName() string {
	return "article_version_files"
}



