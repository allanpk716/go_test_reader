package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/allanpk716/go_test_reader/internal/parser"
	taskpkg "github.com/allanpk716/go_test_reader/internal/task"
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
	if server.taskManager == nil {
		t.Error("Expected task manager to be initialized")
	}
}

// TestMCPServer_HandleUploadTestLog_ValidFile 测试有效文件上传
func TestMCPServer_HandleUploadTestLog_ValidFile(t *testing.T) {
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
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{
			FilePath: tempFile,
		},
	}

	// Act
	result, err := server.handleUploadTestLog(ctx, session, params)

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
	if meta["status"] != "started" {
		t.Errorf("Expected status=started, got %v", meta["status"])
	}
	if meta["task_id"] == nil || meta["task_id"] == "" {
		t.Error("Expected task_id to be set")
	}
}

// TestMCPServer_HandleUploadTestLog_EmptyFilePath 测试空文件路径
func TestMCPServer_HandleUploadTestLog_EmptyFilePath(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{
			FilePath: "",
		},
	}

	// Act
	result, err := server.handleUploadTestLog(ctx, session, params)

	// Assert
	if err == nil {
		t.Error("Expected error for empty file path, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	if !strings.Contains(err.Error(), "file_path parameter is required") {
		t.Errorf("Expected 'file_path parameter is required' error, got %v", err)
	}
}

// TestMCPServer_HandleUploadTestLog_NonExistentFile 测试不存在的文件
func TestMCPServer_HandleUploadTestLog_NonExistentFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{
			FilePath: "/non/existent/file.json",
		},
	}

	// Act
	result, err := server.handleUploadTestLog(ctx, session, params)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error during upload, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 等待任务处理完成
	taskID := result.Meta["task_id"].(string)
	time.Sleep(100 * time.Millisecond)

	// 检查任务状态应该是失败
	task := server.taskManager.GetTask(taskID)
	if task == nil {
		t.Fatal("Expected task to exist")
	}
	if task.Status != taskpkg.StatusFailed {
		t.Errorf("Expected task status to be failed, got %s", task.Status)
	}
}

// TestMCPServer_HandleGetAnalysisResult_ValidTask 测试获取有效任务结果
func TestMCPServer_HandleGetAnalysisResult_ValidTask(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建并设置任务结果
	task := server.taskManager.CreateTask("test-task-1", "/test/path")
	testResult := &parser.TestResult{
		TotalTests:  3,
		PassedTests: 2,
		FailedTests: 1,
	}
	task.SetResult(testResult)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[QueryRequest]{
		Arguments: QueryRequest{
			TaskID: task.ID,
		},
	}

	// Act
	result, err := server.handleGetAnalysisResult(ctx, session, params)

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
	if meta["status"] != string(taskpkg.StatusCompleted) {
		t.Errorf("Expected status=%s, got %v", taskpkg.StatusCompleted, meta["status"])
	}
}

// TestMCPServer_HandleGetAnalysisResult_EmptyTaskID 测试空任务ID
func TestMCPServer_HandleGetAnalysisResult_EmptyTaskID(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[QueryRequest]{
		Arguments: QueryRequest{
			TaskID: "",
		},
	}

	// Act
	result, err := server.handleGetAnalysisResult(ctx, session, params)

	// Assert
	if err == nil {
		t.Error("Expected error for empty task ID, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	if !strings.Contains(err.Error(), "task_id parameter is required") {
		t.Errorf("Expected 'task_id parameter is required' error, got %v", err)
	}
}

// TestMCPServer_HandleGetAnalysisResult_TaskNotFound 测试任务不存在
func TestMCPServer_HandleGetAnalysisResult_TaskNotFound(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[QueryRequest]{
		Arguments: QueryRequest{
			TaskID: "non-existent-task-id",
		},
	}

	// Act
	result, err := server.handleGetAnalysisResult(ctx, session, params)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent task, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got %v", err)
	}
}

// TestMCPServer_HandleTerminateTask_ValidTask 测试终止有效任务
func TestMCPServer_HandleTerminateTask_ValidTask(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	task := server.taskManager.CreateTask("test-task-2", "/test/path")
	task.SetRunning()

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[TerminateRequest]{
		Arguments: TerminateRequest{
			TaskID: task.ID,
		},
	}

	// Act
	result, err := server.handleTerminateTask(ctx, session, params)

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
	if meta["status"] != "terminated" {
		t.Errorf("Expected status=terminated, got %v", meta["status"])
	}
	if meta["task_id"] != task.ID {
		t.Errorf("Expected task_id=%s, got %v", task.ID, meta["task_id"])
	}

	// 验证任务确实被取消
	if task.Status != taskpkg.StatusCanceled {
		t.Errorf("Expected task status to be canceled, got %s", task.Status)
	}
}

// TestMCPServer_HandleTerminateTask_EmptyTaskID 测试空任务ID终止
func TestMCPServer_HandleTerminateTask_EmptyTaskID(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[TerminateRequest]{
		Arguments: TerminateRequest{
			TaskID: "",
		},
	}

	// Act
	result, err := server.handleTerminateTask(ctx, session, params)

	// Assert
	if err == nil {
		t.Error("Expected error for empty task ID, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	if !strings.Contains(err.Error(), "task_id parameter is required") {
		t.Errorf("Expected 'task_id parameter is required' error, got %v", err)
	}
}

// TestMCPServer_HandleGetTestDetails_ValidTest 测试获取有效测试详情
func TestMCPServer_HandleGetTestDetails_ValidTest(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	task := server.taskManager.CreateTask("test-task-3", "/test/path")
	testDetails := map[string]*parser.TestDetail{
		"TestExample": {
			Status:  "pass",
			Output:  "test output",
			Error:   "",
			Elapsed: 0.001,
		},
	}
	testResult := &parser.TestResult{
		TestDetails: testDetails,
	}
	task.SetResult(testResult)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[TestDetailsRequest]{
		Arguments: TestDetailsRequest{
			TaskID:   task.ID,
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

// TestMCPServer_HandleGetTestDetails_EmptyParameters 测试空参数
func TestMCPServer_HandleGetTestDetails_EmptyParameters(t *testing.T) {
	tests := []struct {
		name     string
		taskID   string
		testName string
		expectedError string
	}{
		{
			name:          "Empty task ID",
			taskID:        "",
			testName:      "TestExample",
			expectedError: "task_id parameter is required",
		},
		{
			name:          "Empty test name",
			taskID:        "valid-task-id",
			testName:      "",
			expectedError: "test_name parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server, err := NewMCPServer()
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			ctx := context.Background()
			session := &mcp.ServerSession{}
			params := &mcp.CallToolParamsFor[TestDetailsRequest]{
				Arguments: TestDetailsRequest{
					TaskID:   tt.taskID,
					TestName: tt.testName,
				},
			}

			// Act
			result, err := server.handleGetTestDetails(ctx, session, params)

			// Assert
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if result != nil {
				t.Errorf("Expected nil result, got %v", result)
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected '%s' error, got %v", tt.expectedError, err)
			}
		})
	}
}

// TestMCPServer_ProcessTestLog_ValidFile 测试处理有效测试日志文件
func TestMCPServer_ProcessTestLog_ValidFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tempFile := createTempTestFile(t, validTestLogContent())
	defer os.Remove(tempFile)

	task := server.taskManager.CreateTask("test-task-4", tempFile)
	ctx := context.Background()

	// Act
	server.processTestLog(ctx, task)

	// Assert
	if task.Status != taskpkg.StatusCompleted {
		t.Errorf("Expected task status to be completed, got %s", task.Status)
	}
	if task.Result == nil {
		t.Error("Expected task result to be set")
	}
	if task.Error != nil {
		t.Errorf("Expected no error, got %v", task.Error)
	}
}

// TestMCPServer_ProcessTestLog_InvalidFile 测试处理无效文件
func TestMCPServer_ProcessTestLog_InvalidFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	task := server.taskManager.CreateTask("test-task-5", "/non/existent/file.json")
	ctx := context.Background()

	// Act
	server.processTestLog(ctx, task)

	// Assert
	if task.Status != taskpkg.StatusFailed {
		t.Errorf("Expected task status to be failed, got %s", task.Status)
	}
	if task.Error == nil {
		t.Error("Expected error to be set")
	}
	if task.Result != nil {
		t.Errorf("Expected no result, got %v", task.Result)
	}
}

// 辅助函数
func createTempTestFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tempFile
}

func validTestLogContent() string {
	return `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"=== RUN   TestExample1\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}
{"Time":"2024-01-15T10:30:03Z","Action":"run","Package":"example/pkg","Test":"TestExample2"}
{"Time":"2024-01-15T10:30:04Z","Action":"output","Package":"example/pkg","Test":"TestExample2","Output":"=== RUN   TestExample2\n"}
{"Time":"2024-01-15T10:30:05Z","Action":"output","Package":"example/pkg","Test":"TestExample2","Output":"FAIL: Expected 5, got 3\n"}
{"Time":"2024-01-15T10:30:06Z","Action":"fail","Package":"example/pkg","Test":"TestExample2","Elapsed":0.002}`
}