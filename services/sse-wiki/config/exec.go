package config

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func initSwaggerFile() error {
	log.Println("Start initing swagger file")
	// Build the command and run it with a timeout so it won't hang the process.
	args := []string{
		"run",
		"github.com/swaggo/swag/cmd/swag@latest",
		"init",
		"-g",
		"cmd/server/main.go",
		"-o",
		"docs",
		"--parseDependency",
		"--parseInternal",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run swag init: %w; stdout: %s; stderr: %s", err, strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()))
	}

	log.Printf("swag init completed\n")
	return nil
}

// 通过命令初始化
func InitProgram() {
	err := initSwaggerFile()
	if err != nil {
		log.Panic("Fail to init swagger file")
	}
}
