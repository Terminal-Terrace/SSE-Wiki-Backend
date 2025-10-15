package article

import (
	"encoding/json"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type MergeService struct{}

// MergeResult 3路合并结果
type MergeResult struct {
	HasConflict           bool   `json:"has_conflict"`
	MergedContent         string `json:"merged_content"`
	ConflictMarkedContent string `json:"conflict_marked_content"`
	ConflictDetails       string `json:"conflict_details"` // JSON格式
}

func NewMergeService() *MergeService {
	return &MergeService{}
}

// ThreeWayMerge 执行3路合并
// base: 共同祖先版本
// theirs: 提交者的修改
// ours: 当前线上版本
func (s *MergeService) ThreeWayMerge(base, theirs, ours string) MergeResult {
	dmp := diffmatchpatch.New()

	// 1. 计算 base -> theirs 的patch
	theirDiffs := dmp.DiffMain(base, theirs, false)
	theirPatches := dmp.PatchMake(base, theirDiffs)

	// 2. 计算 base -> ours 的patch
	ourDiffs := dmp.DiffMain(base, ours, false)
	ourPatches := dmp.PatchMake(base, ourDiffs)

	// 3. 尝试将两组patch都应用到base上
	mergedTheirs, _ := dmp.PatchApply(theirPatches, base)
	mergedBoth, conflicts := dmp.PatchApply(ourPatches, mergedTheirs)

	// 4. 检查是否有冲突
	hasConflict := false
	for _, success := range conflicts {
		if !success {
			hasConflict = true
			break
		}
	}

	if hasConflict {
		// 生成带冲突标记的内容
		conflictMarked := s.generateConflictMarkers(base, theirs, ours, theirDiffs, ourDiffs)
		conflictDetails := s.serializeConflictDetails(theirDiffs, ourDiffs)

		return MergeResult{
			HasConflict:           true,
			ConflictMarkedContent: conflictMarked,
			ConflictDetails:       conflictDetails,
		}
	}

	return MergeResult{
		HasConflict:   false,
		MergedContent: mergedBoth,
	}
}

// generateConflictMarkers 生成带冲突标记的内容
func (s *MergeService) generateConflictMarkers(base, theirs, ours string, theirDiffs, ourDiffs []diffmatchpatch.Diff) string {
	var result strings.Builder

	result.WriteString("<<<<<<< THEIRS (提交者的修改)\n")
	result.WriteString(theirs)
	result.WriteString("\n=======\n")
	result.WriteString(ours)
	result.WriteString("\n>>>>>>> OURS (当前线上版本)\n")

	return result.String()
}

// serializeConflictDetails 序列化冲突详情（JSON格式）
func (s *MergeService) serializeConflictDetails(theirDiffs, ourDiffs []diffmatchpatch.Diff) string {
	details := map[string]interface{}{
		"theirChanges": len(theirDiffs),
		"ourChanges":   len(ourDiffs),
	}

	jsonData, _ := json.Marshal(details)
	return string(jsonData)
}
