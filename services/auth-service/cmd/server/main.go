package main

import (
	"terminal-terrace/auth-service/internal/route"
	"terminal-terrace/auth-service/config"
)

func main() {
	config.Load("config.yaml")
	r := route.SetupRouter()

	r.Run(":8080")
}