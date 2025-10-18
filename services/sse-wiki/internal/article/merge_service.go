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
// TODO: 合并算法优化 - 当前实现是简化版的字符串级合并，未来可以实现更智能的行级合并或语义级合并
func (s *MergeService) ThreeWayMerge(base, theirs, ours string) MergeResult {
	// 1. 检查是否有实质性的修改冲突
	// 如果 theirs 和 ours 都不等于 base，且它们也不相等，说明有冲突
	theirsChanged := theirs != base
	oursChanged := ours != base

	if theirsChanged && oursChanged && theirs != ours {
		// 两者都改了，且改的不一样 → 冲突！
		dmp := diffmatchpatch.New()
		theirDiffs := dmp.DiffMain(base, theirs, false)
		ourDiffs := dmp.DiffMain(base, ours, false)

		conflictMarked := s.generateConflictMarkers(base, theirs, ours, theirDiffs, ourDiffs)
		conflictDetails := s.serializeConflictDetails(theirDiffs, ourDiffs)

		return MergeResult{
			HasConflict:           true,
			ConflictMarkedContent: conflictMarked,
			ConflictDetails:       conflictDetails,
		}
	}

	// 4. 尝试自动合并（只在其中一方修改时）
	if !theirsChanged && oursChanged {
		// 只有 ours 改了，使用 ours
		return MergeResult{
			HasConflict:   false,
			MergedContent: ours,
		}
	}

	if theirsChanged && !oursChanged {
		// 只有 theirs 改了，使用 theirs
		return MergeResult{
			HasConflict:   false,
			MergedContent: theirs,
		}
	}

	// 5. 两者都没改，或者改的一样
	if theirs == ours {
		return MergeResult{
			HasConflict:   false,
			MergedContent: theirs,
		}
	}

	// 默认返回 base（理论上不会到这里）
	return MergeResult{
		HasConflict:   false,
		MergedContent: base,
	}
}

// generateConflictMarkers 生成带冲突标记的内容
// 使用标准的两路冲突标记格式（Git 风格）
func (s *MergeService) generateConflictMarkers(base, theirs, ours string, theirDiffs, ourDiffs []diffmatchpatch.Diff) string {
	var result strings.Builder

	// 写入冲突标记开始
	result.WriteString("<<<<<<< THEIRS (提交者的修改)\n")

	// 写入 theirs 内容（提交者的修改）
	result.WriteString(theirs)
	if !strings.HasSuffix(theirs, "\n") {
		result.WriteString("\n")
	}

	// 写入分隔符
	result.WriteString("=======\n")

	// 写入 ours 内容（当前线上版本）
	result.WriteString(ours)
	if !strings.HasSuffix(ours, "\n") {
		result.WriteString("\n")
	}

	// 写入冲突标记结束
	result.WriteString(">>>>>>> OURS (当前线上版本)\n")

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
