package main

import (
	"terminal-terrace/sse-wiki/internal/route"
	"terminal-terrace/sse-wiki/config"
)

func main() {
	config.Load("config.yaml")
	r := route.SetupRouter()

	r.Run(":8080")
}