module terminal-terrace/template

go 1.25.0

require (
	github.com/gin-contrib/cors v1.7.3
	github.com/gin-gonic/gin v1.10.0
	github.com/joho/godotenv v1.5.1
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/env v1.0.0
	github.com/knadh/koanf/providers/file v1.1.2
	github.com/knadh/koanf/v2 v2.1.2
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	github.com/swaggo/swag v1.16.6
	gorm.io/gorm v1.31.0
	terminal-terrace/database v0.0.0
	terminal-terrace/response v0.0.0
)

replace (
	terminal-terrace/database => ../../packages/database
	terminal-terrace/response => ../../packages/response
)
