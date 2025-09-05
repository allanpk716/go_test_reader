package server

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCPServer_HandleAnalyzeTestLog_InvalidFile 测试无效文件分析
func TestMCPServer_HandleAnalyzeTestLog_InvalidFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: AnalyzeTestLogRequest{
			FilePath: "/nonexistent/file.json",
		},
	}

	// Act
	_, err = server.handleAnalyzeTestLog(ctx, session, params)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "cannot find") {
		t.Errorf("Expected file not found error, got %v", err)
	}
}

// TestMCPServer_HandleAnalyzeTestLog_CorruptedJSON 测试损坏的JSON文件
func TestMCPServer_HandleAnalyzeTestLog_CorruptedJSON(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建损坏的JSON文件
	corruptedContent := `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"example","Test":"TestExample"}
{CORRUPTED JSON LINE}
{"Time":"2023-01-01T00:00:01Z","Action":"pass","Package":"example","Test":"TestExample","Elapsed":0.1}`
	tempFile := createTempTestFileForFailure(t, corruptedContent)
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
	// 解析器应该跳过无效行并处理有效行
	if err != nil {
		t.Fatalf("Expected no error (parser should skip invalid lines), got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 应该至少解析出一个测试
	meta := result.Meta
	if meta["total_tests"].(int) < 1 {
		t.Errorf("Expected at least 1 test to be parsed, got %v", meta["total_tests"])
	}
}

// TestMCPServer_HandleAnalyzeTestLog_EmptyFile 测试空文件
func TestMCPServer_HandleAnalyzeTestLog_EmptyFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tempFile := createTempTestFileForFailure(t, "")
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: AnalyzeTestLogRequest{
			FilePath: tempFile,
		},
	}

	// Act
	_, err = server.handleAnalyzeTestLog(ctx, session, params)

	// Assert
	// 空文件应该返回错误，因为不是有效的go test输出
	if err == nil {
		t.Fatal("Expected error for empty file")
	}
	if !strings.Contains(err.Error(), "valid go test output") {
		t.Errorf("Expected 'valid go test output' error, got %v", err)
	}
}

// TestMCPServer_HandleGetTestDetails_NonexistentFile 测试不存在的文件
func TestMCPServer_HandleGetTestDetails_NonexistentFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: GetTestDetailsRequest{
			FilePath: "/nonexistent/file.json",
			TestName: "TestExample",
		},
	}

	// Act
	_, err = server.handleGetTestDetails(ctx, session, params)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

// TestMCPServer_HandleGetTestDetails_TestNotFound 测试不存在的测试名称
func TestMCPServer_HandleGetTestDetails_TestNotFound(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tempFile := createTempTestFileForFailure(t, validTestLogContent())
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: GetTestDetailsRequest{
			FilePath: tempFile,
			TestName: "NonexistentTest",
		},
	}

	// Act
	_, err = server.handleGetTestDetails(ctx, session, params)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent test")
	}
	if !strings.Contains(err.Error(), "test not found") {
		t.Errorf("Expected 'test not found' error, got %v", err)
	}
}

// TestMCPServer_HandleGetTestDetails_EmptyTestName 测试空测试名称
func TestMCPServer_HandleGetTestDetails_EmptyTestName(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tempFile := createTempTestFileForFailure(t, validTestLogContent())
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: GetTestDetailsRequest{
			FilePath: tempFile,
			TestName: "",
		},
	}

	// Act
	_, err = server.handleGetTestDetails(ctx, session, params)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty test name")
	}
}

func createTempTestFileForFailure(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test_failure_*.json")
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		panic(err)
	}
	defer tempFile.Close()

	_, err = tempFile.WriteString(content)
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		panic(err)
	}

	return tempFile.Name()
}

// validTestLogContent 函数在 server_test.go 中定义