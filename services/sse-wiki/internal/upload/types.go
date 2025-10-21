package upload

type InitUploadRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	FileSize    int64  `json:"fileSize" binding:"required,gt=0"`
	FileHash    string `json:"fileHash" binding:"required,len=64"`
	TotalChunks int    `json:"totalChunks" binding:"required,gt=0"`
	MimeType    string `json:"mimeType" binding:"required"`
}

type InitUploadResponse struct {
	Exists   bool   `json:"exists"`
	UploadID string `json:"uploadId,omitempty"`
	FileID   uint   `json:"fileId,omitempty"`
	FileURL  string `json:"fileUrl,omitempty"`
}

type CompleteUploadRequest struct {
	UploadID string `json:"uploadId" binding:"required"`
}

type UploadSession struct {
	UploadID       string `json:"uploadId"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	FileHash       string `json:"fileHash"`
	TotalChunks    int    `json:"totalChunks"`
	UploadedChunks []bool `json:"uploadedChunks"`
	MimeType       string `json:"mimeType"`
}


