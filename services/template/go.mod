module terminal-terrace/template

go 1.25.0

require (
	github.com/joho/godotenv v1.5.1
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/env v1.0.0
	github.com/knadh/koanf/providers/file v1.1.2
	github.com/knadh/koanf/v2 v2.1.2
	google.golang.org/grpc v1.72.0
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.31.0
	terminal-terrace/database v0.0.0
	terminal-terrace/response v0.0.0
)

replace (
	terminal-terrace/database => ../../packages/database
	terminal-terrace/response => ../../packages/response
)
