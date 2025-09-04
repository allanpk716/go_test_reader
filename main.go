package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/allanpk716/go_test_reader/internal/server"
)

func main() {
	ctx := context.Background()

	// 创建 MCP 服务器
	mcpServer, err := server.NewMCPServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// 启动服务器
	if err := mcpServer.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
