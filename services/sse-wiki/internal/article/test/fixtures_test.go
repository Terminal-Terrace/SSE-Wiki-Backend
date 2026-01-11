package article_test

import (
	"strings"
	"testing"
	"time"

	articlePkg "terminal-terrace/sse-wiki/internal/article"
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/model/module"
	"terminal-terrace/sse-wiki/internal/model/user"
	"terminal-terrace/sse-wiki/internal/testutils"
	"gorm.io/gorm"
)

// setupArticleService 创建 ArticleService 实例用于测试
func setupArticleService(t *testing.T) (*articlePkg.ArticleService, *gorm.DB) {
	db := testutils.SetupTestDB(t)

	articleRepo := articlePkg.NewArticleRepository(db)
	versionRepo := articlePkg.NewVersionRepository(db)
	submissionRepo := articlePkg.NewSubmissionRepository(db)
	tagRepo := articlePkg.NewTagRepository(db)
	mergeService := articlePkg.NewMergeService()

	service := articlePkg.NewArticleService(articleRepo, versionRepo, submissionRepo, tagRepo, mergeService)
	return service, db
}

// ArticleTestFixture 共享测试数据结构
type ArticleTestFixture struct {
	DB      *gorm.DB
	Service *articlePkg.ArticleService

	// Users
	Author       *user.User
	AdminUser    *user.User
	ModeratorUser *user.User
	RegularUser  *user.User
	GlobalAdmin  *user.User

	// Module and Article
	TestModule  *module.Module
	TestArticle *article.Article
	BaseVersion *article.ArticleVersion
}

// createArticleFixture 创建完整的文章测试fixture
func createArticleFixture(t *testing.T) *ArticleTestFixture {
	service, db := setupArticleService(t)

	author := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Initial content",
		CommitMessage: "Initial commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	if err := db.Create(baseVersion).Error; err != nil {
		t.Fatalf("Failed to create base version: %v", err)
	}
	testArticle.CurrentVersionID = &baseVersion.ID
	if err := db.Save(testArticle).Error; err != nil {
		t.Fatalf("Failed to update article: %v", err)
	}

	// 添加作者为 admin 协作者
	db.Table("article_collaborators").FirstOrCreate(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    author.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})

	return &ArticleTestFixture{
		DB:           db,
		Service:      service,
		Author:       author,
		AdminUser:    adminUser,
		ModeratorUser: moderatorUser,
		RegularUser:  regularUser,
		GlobalAdmin:  globalAdmin,
		TestModule:   testModule,
		TestArticle:  testArticle,
		BaseVersion:  baseVersion,
	}
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// boolPtr 返回布尔指针
func boolPtr(b bool) *bool {
	return &b
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ============================================================================
// 测试数据常量 - 提取重复的HTML内容以减少代码重复
// ============================================================================

// SSE-Wiki使用指南相关测试内容
const (
	// BaseContentSSEWiki SSE-Wiki使用指南基础版本
	BaseContentSSEWiki = `<h1>SSE-Wiki 使用指南</h1>
<p>这是一个协作式Wiki系统。</p>
<p>支持多人协作编辑。</p>`

	// UserAContentSSEWiki 用户A的修改：第一段添加"支持实时协作"
	UserAContentSSEWiki = `<h1>SSE-Wiki 使用指南</h1>
<p>这是一个协作式Wiki系统，支持实时协作。</p>
<p>支持多人协作编辑。</p>`

	// UserBContentSSEWiki 用户B的修改：第二段添加"支持版本管理"
	UserBContentSSEWiki = `<h1>SSE-Wiki 使用指南</h1>
<p>这是一个协作式Wiki系统。</p>
<p>支持多人协作编辑，支持版本管理。</p>`

	// UserCContentSSEWiki 用户C的修改：修改标题
	UserCContentSSEWiki = `<h1>SSE-Wiki 使用指南（更新）</h1>
<p>这是一个协作式Wiki系统。</p>
<p>支持多人协作编辑。</p>`

	// UserDContentSSEWiki 用户D的修改：第一段添加"支持Markdown"
	UserDContentSSEWiki = `<h1>SSE-Wiki 使用指南</h1>
<p>这是一个协作式Wiki系统，支持Markdown。</p>
<p>支持多人协作编辑。</p>`

	// UserEContentSSEWiki 用户E的修改：第二段添加"支持权限管理"
	UserEContentSSEWiki = `<h1>SSE-Wiki 使用指南</h1>
<p>这是一个协作式Wiki系统。</p>
<p>支持多人协作编辑，支持权限管理。</p>`
)

// Go语言教程相关测试内容
const (
	// BaseContentGoTutorial Go语言教程基础版本
	BaseContentGoTutorial = `<h1>Go语言教程</h1>
<section>
  <h2>第一章：基础语法</h2>
  <p>Go语言是Google开发的编程语言。</p>
</section>
<section>
  <h2>第二章：并发编程</h2>
  <p>Go语言支持goroutine。</p>
</section>`

	// UserAContentGoTutorial 用户A的修改：第一章添加"语法简洁"
	UserAContentGoTutorial = `<h1>Go语言教程</h1>
<section>
  <h2>第一章：基础语法</h2>
  <p>Go语言是Google开发的编程语言，语法简洁。</p>
</section>
<section>
  <h2>第二章：并发编程</h2>
  <p>Go语言支持goroutine。</p>
</section>`

	// UserBContentGoTutorial 用户B的修改：基于版本2，第二章添加"和channel"
	UserBContentGoTutorial = `<h1>Go语言教程</h1>
<section>
  <h2>第一章：基础语法</h2>
  <p>Go语言是Google开发的编程语言，语法简洁。</p>
</section>
<section>
  <h2>第二章：并发编程</h2>
  <p>Go语言支持goroutine和channel。</p>
</section>`
)

// 简单内容测试数据（用于冲突解决等场景）
const (
	// BaseContentSimple 简单原始内容
	BaseContentSimple = `<p>原始内容</p>`

	// UserAContentSimple 用户A的简单修改
	UserAContentSimple = `<p>用户A的修改</p>`

	// UserBContentSimple 用户B的简单修改
	UserBContentSimple = `<p>用户B的修改</p>`

	// ResolvedContentSimple 手动解决的合并内容
	ResolvedContentSimple = `<p>手动解决的内容：A和B的合并</p>`

	// UserCContentSimple 用户C基于解决后的内容
	UserCContentSimple = `<p>手动解决的内容：A和B的合并
用户C的新修改</p>`
)

