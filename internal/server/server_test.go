package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestNewMCPServer 测试MCP服务器创建
func TestNewMCPServer(t *testing.T) {
	// Act
	server, err := NewMCPServer()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if server == nil {
		t.Fatal("Expected server, got nil")
	}
	if server.server == nil {
		t.Error("Expected MCP server to be initialized")
	}
}

// TestMCPServer_HandleAnalyzeTestLog_ValidFile 测试有效文件分析
func TestMCPServer_HandleAnalyzeTestLog_ValidFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建临时测试文件
	tempFile := createTempTestFile(t, validTestLogContent())
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: AnalyzeTestLogRequest{
			FilePath: tempFile,
		},
	}

	// Act
	result, err := server.handleAnalyzeTestLog(ctx, session, params)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Meta == nil {
		t.Fatal("Expected meta, got nil")
	}

	meta := result.Meta
	if meta["all_tests_passed"] != true {
		t.Errorf("Expected all_tests_passed=true, got %v", meta["all_tests_passed"])
	}
	if meta["total_tests"] != 2 {
		t.Errorf("Expected total_tests=2, got %v", meta["total_tests"])
	}
}

// TestMCPServer_HandleAnalyzeTestLog_EmptyFilePath 测试空文件路径
func TestMCPServer_HandleAnalyzeTestLog_EmptyFilePath(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: AnalyzeTestLogRequest{
			FilePath: "",
		},
	}

	// Act
	_, err = server.handleAnalyzeTestLog(ctx, session, params)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty file path")
	}
}

// TestMCPServer_HandleGetTestDetails_ValidTest 测试获取测试详情
func TestMCPServer_HandleGetTestDetails_ValidTest(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建临时测试文件
	tempFile := createTempTestFile(t, validTestLogContent())
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: GetTestDetailsRequest{
			FilePath: tempFile,
			TestName: "TestExample",
		},
	}

	// Act
	result, err := server.handleGetTestDetails(ctx, session, params)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Meta == nil {
		t.Fatal("Expected meta, got nil")
	}

	meta := result.Meta
	if meta["test_name"] != "TestExample" {
		t.Errorf("Expected test_name=TestExample, got %v", meta["test_name"])
	}
	if meta["status"] != "pass" {
		t.Errorf("Expected status=pass, got %v", meta["status"])
	}
}

func createTempTestFile(t *testing.T, content string) string {
	tempFile := filepath.Join(t.TempDir(), "test.json")
	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tempFile
}

func validTestLogContent() string {
	return `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"example","Test":"TestExample"}
{"Time":"2023-01-01T00:00:01Z","Action":"pass","Package":"example","Test":"TestExample","Elapsed":0.1}
{"Time":"2023-01-01T00:00:02Z","Action":"run","Package":"example","Test":"TestAnother"}
{"Time":"2023-01-01T00:00:03Z","Action":"pass","Package":"example","Test":"TestAnother","Elapsed":0.2}`
}