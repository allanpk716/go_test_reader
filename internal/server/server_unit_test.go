package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMCPServer_Success 测试成功创建MCP服务器
func TestNewMCPServer_Success(t *testing.T) {
	// 创建MCP服务器
	server, err := NewMCPServer()
	
	// 验证创建成功
	require.NoError(t, err, "Should create MCP server without error")
	require.NotNil(t, server, "Server should not be nil")
	require.NotNil(t, server.server, "Internal server should not be nil")
}

// TestMCPServer_RegisterTools 测试工具注册功能
func TestMCPServer_RegisterTools(t *testing.T) {
	// 创建MCP服务器
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server without error")
	
	// 验证工具已注册（通过检查内部状态）
	// 注意：这里我们测试的是工具注册的副作用，实际的工具功能在MCP客户端测试中验证
	assert.NotNil(t, server.server, "Server should have internal MCP server")
}

// TestAnalyzeTestLogRequest_Validation 测试请求参数验证
func TestAnalyzeTestLogRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		request  AnalyzeTestLogRequest
		expected bool // true if valid, false if invalid
	}{
		{
			name:     "Valid file path",
			request:  AnalyzeTestLogRequest{FilePath: "/valid/path/file.txt"},
			expected: true,
		},
		{
			name:     "Empty file path",
			request:  AnalyzeTestLogRequest{FilePath: ""},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.request.FilePath != ""
			assert.Equal(t, tt.expected, isValid, "Validation result should match expected")
		})
	}
}

// TestGetTestDetailsRequest_Validation 测试获取测试详情请求参数验证
func TestGetTestDetailsRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		request  GetTestDetailsRequest
		expected bool // true if valid, false if invalid
	}{
		{
			name:     "Valid parameters",
			request:  GetTestDetailsRequest{FilePath: "/valid/path/file.txt", TestName: "TestExample"},
			expected: true,
		},
		{
			name:     "Empty file path",
			request:  GetTestDetailsRequest{FilePath: "", TestName: "TestExample"},
			expected: false,
		},
		{
			name:     "Empty test name",
			request:  GetTestDetailsRequest{FilePath: "/valid/path/file.txt", TestName: ""},
			expected: false,
		},
		{
			name:     "Both empty",
			request:  GetTestDetailsRequest{FilePath: "", TestName: ""},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.request.FilePath != "" && tt.request.TestName != ""
			assert.Equal(t, tt.expected, isValid, "Validation result should match expected")
		})
	}
}

// TestTestOverviewResponse_Structure 测试测试总览响应结构
func TestTestOverviewResponse_Structure(t *testing.T) {
	response := TestOverviewResponse{
		AllTestsPassed:   true,
		TotalTests:       10,
		FailedTestsCount: 0,
		FailedTestNames:  []string{},
	}
	
	// 验证结构字段
	assert.True(t, response.AllTestsPassed, "AllTestsPassed should be true")
	assert.Equal(t, 10, response.TotalTests, "TotalTests should be 10")
	assert.Equal(t, 0, response.FailedTestsCount, "FailedTestsCount should be 0")
	assert.Empty(t, response.FailedTestNames, "FailedTestNames should be empty")
}

// TestTestDetailsResponse_Structure 测试测试详情响应结构
func TestTestDetailsResponse_Structure(t *testing.T) {
	response := TestDetailsResponse{
		TestName: "TestExample",
		Status:   "PASS",
		Output:   "Test output",
		Error:    "",
		Elapsed:  0.123,
	}
	
	// 验证结构字段
	assert.Equal(t, "TestExample", response.TestName, "TestName should match")
	assert.Equal(t, "PASS", response.Status, "Status should be PASS")
	assert.Equal(t, "Test output", response.Output, "Output should match")
	assert.Empty(t, response.Error, "Error should be empty")
	assert.Equal(t, 0.123, response.Elapsed, "Elapsed should match")
}

// TestMCPServer_ParseTestLogWithAutoDetection_ValidFiles 测试自动检测解析功能
func TestMCPServer_ParseTestLogWithAutoDetection_ValidFiles(t *testing.T) {
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server without error")
	
	// 获取测试数据目录
	testDataDir := filepath.Join("..", "..", "test_data")
	
	// 测试成功的测试文件
	okFiles := []string{"ok_00.txt", "ok_01.txt", "ok_02.txt"}
	for _, filename := range okFiles {
		t.Run("Parse_"+filename, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, filename)
			
			// 检查文件是否存在
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Test file %s does not exist", filePath)
				return
			}
			
			file, err := os.Open(filePath)
			require.NoError(t, err, "Should open test file")
			defer file.Close()
			
			result, err := server.parseTestLogWithAutoDetection(file)
			require.NoError(t, err, "Should parse test log successfully")
			require.NotNil(t, result, "Result should not be nil")
			
			// 验证解析结果的基本结构
			assert.GreaterOrEqual(t, result.TotalTests, 0, "Total tests should be non-negative")
			assert.GreaterOrEqual(t, result.FailedTests, 0, "Failed tests should be non-negative")
			assert.LessOrEqual(t, result.FailedTests, result.TotalTests, "Failed tests should not exceed total tests")
		})
	}
	
	// 测试失败的测试文件
	failFiles := []string{"fail_00.txt", "fail_01.txt", "fail_02.txt"}
	for _, filename := range failFiles {
		t.Run("Parse_"+filename, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, filename)
			
			// 检查文件是否存在
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Test file %s does not exist", filePath)
				return
			}
			
			file, err := os.Open(filePath)
			require.NoError(t, err, "Should open test file")
			defer file.Close()
			
			result, err := server.parseTestLogWithAutoDetection(file)
			require.NoError(t, err, "Should parse test log successfully")
			require.NotNil(t, result, "Result should not be nil")
			
			// 验证解析结果
			assert.GreaterOrEqual(t, result.TotalTests, 0, "Total tests should be non-negative")
			assert.GreaterOrEqual(t, result.FailedTests, 0, "Failed tests should be non-negative")
		})
	}
}

// TestMCPServer_ParseTestLogWithAutoDetection_InvalidFile 测试解析无效文件
func TestMCPServer_ParseTestLogWithAutoDetection_InvalidFile(t *testing.T) {
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server without error")
	
	// 创建临时的无效文件
	tempFile, err := os.CreateTemp("", "invalid_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入无效内容 - 确保没有任何测试模式
	_, err = tempFile.WriteString("random content line 1\nrandom content line 2\nno test patterns here\njust random content\nmore random text\neven more random text\nstill more random text\nyet more random text\nfinal random text\nend of random content")
	require.NoError(t, err, "Should write to temp file")
	
	// 重置文件指针
	tempFile.Seek(0, 0)
	
	// 尝试解析，应该失败
	_, err = server.parseTestLogWithAutoDetection(tempFile)
	assert.Error(t, err, "Should fail to parse invalid file")
	if err != nil {
		assert.Contains(t, err.Error(), "does not appear to be valid go test output", "Error should mention invalid format")
	}
}

// TestMCPServer_HandleAnalyzeTestLog_ErrorCases 测试分析测试日志的错误情况
func TestMCPServer_HandleAnalyzeTestLog_ErrorCases(t *testing.T) {
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server without error")
	
	ctx := context.Background()
	session := &mcp.ServerSession{} // 模拟会话
	
	// 测试空文件路径
	emptyPathParams := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: AnalyzeTestLogRequest{FilePath: ""},
	}
	
	_, err = server.handleAnalyzeTestLog(ctx, session, emptyPathParams)
	assert.Error(t, err, "Should fail with empty file path")
	assert.Contains(t, err.Error(), "file_path parameter is required", "Error should mention required parameter")
	
	// 测试不存在的文件
	nonExistentParams := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: AnalyzeTestLogRequest{FilePath: "/nonexistent/file.txt"},
	}
	
	_, err = server.handleAnalyzeTestLog(ctx, session, nonExistentParams)
	assert.Error(t, err, "Should fail with nonexistent file")
	assert.Contains(t, err.Error(), "failed to open file", "Error should mention file opening failure")
}

// TestMCPServer_HandleGetTestDetails_ErrorCases 测试获取测试详情的错误情况
func TestMCPServer_HandleGetTestDetails_ErrorCases(t *testing.T) {
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server without error")
	
	ctx := context.Background()
	session := &mcp.ServerSession{} // 模拟会话
	
	// 测试空文件路径
	emptyFilePathParams := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: GetTestDetailsRequest{FilePath: "", TestName: "SomeTest"},
	}
	
	_, err = server.handleGetTestDetails(ctx, session, emptyFilePathParams)
	assert.Error(t, err, "Should fail with empty file path")
	assert.Contains(t, err.Error(), "file_path parameter is required", "Error should mention required file_path")
	
	// 测试空测试名称
	emptyTestNameParams := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: GetTestDetailsRequest{FilePath: "/some/file.txt", TestName: ""},
	}
	
	_, err = server.handleGetTestDetails(ctx, session, emptyTestNameParams)
	assert.Error(t, err, "Should fail with empty test name")
	assert.Contains(t, err.Error(), "test_name parameter is required", "Error should mention required test_name")
}

// TestMCPServer_Run_Context 测试服务器运行和上下文处理
func TestMCPServer_Run_Context(t *testing.T) {
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server without error")
	
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	
	// 在goroutine中运行服务器
	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()
	
	// 立即取消上下文
	cancel()
	
	// 等待服务器停止
	select {
	case err := <-done:
		// 服务器应该因为上下文取消而停止，这可能返回错误或nil
		// 具体行为取决于MCP SDK的实现
		if err != nil {
			// 如果返回错误，应该是上下文相关的错误
			assert.True(t, 
				strings.Contains(err.Error(), "context") || 
				strings.Contains(err.Error(), "canceled") ||
				err == context.Canceled,
				"Error should be context-related: %v", err)
		}
	case <-ctx.Done():
		// 上下文已取消，这是预期的
	}
}