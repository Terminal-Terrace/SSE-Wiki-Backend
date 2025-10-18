package article

type MergeService struct{}

// MergeResult 3路合并结果
// 说明：
// - HasConflict: 是否有冲突（后端检测）
// - MergedContent: 自动合并后的内容（仅当无冲突时有值）
// 注意：冲突标记的生成已移至前端，后端只返回三方原始内容（base, their, our）
type MergeResult struct {
	HasConflict   bool   `json:"has_conflict"`
	MergedContent string `json:"merged_content"` // 仅当 HasConflict=false 时有值
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
		return MergeResult{HasConflict: true}
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

	// 默认返回 theirs（防止为空）
	return MergeResult{
		HasConflict:   false,
		MergedContent: theirs,
	}
}
