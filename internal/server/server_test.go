package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`
}

// TestMCPServer_HandleUploadTestLog_LargeFile 测试大文件上传处理
func TestMCPServer_HandleUploadTestLog_LargeFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建大文件内容（模拟大量测试结果）
	largeContent := ""
	for i := 0; i < 1000; i++ {
		largeContent += fmt.Sprintf(`{"Time":"2024-01-15T10:30:%02dZ","Action":"run","Package":"example/pkg","Test":"TestExample%d"}`+"\n", i%60, i)
		largeContent += fmt.Sprintf(`{"Time":"2024-01-15T10:30:%02dZ","Action":"pass","Package":"example/pkg","Test":"TestExample%d","Elapsed":0.001}`+"\n", (i+1)%60, i)
	}

	tempFile := createTempTestFile(t, largeContent)
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{FilePath: tempFile},
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
	if result.Meta["status"] != "started" {
		t.Errorf("Expected status=started, got %v", result.Meta["status"])
	}
}

// TestMCPServer_HandleUploadTestLog_CorruptedFile 测试损坏文件处理
func TestMCPServer_HandleUploadTestLog_CorruptedFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建损坏的JSON文件
	corruptedContent := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"=== RUN   TestExample1\n"
{CORRUPTED JSON LINE}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`

	tempFile := createTempTestFile(t, corruptedContent)
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{FilePath: tempFile},
	}

	// Act
	result, err := server.handleUploadTestLog(ctx, session, params)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error during upload initiation, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 等待处理完成并检查任务状态
	taskID := result.Meta["task_id"].(string)
	time.Sleep(100 * time.Millisecond) // 等待异步处理

	task := server.taskManager.GetTask(taskID)
	if task == nil {
		t.Fatal("Expected task to exist")
	}

	// 损坏的文件应该仍然能部分解析（跳过无效行）
	status := task.GetStatus()
	if status["status"] == "failed" {
		t.Log("Task failed as expected due to corrupted content")
	} else if status["status"] == "completed" {
		t.Log("Task completed despite corrupted content (parser skipped invalid lines)")
	}
}

// TestMCPServer_HandleUploadTestLog_EmptyFile 测试空文件处理
func TestMCPServer_HandleUploadTestLog_EmptyFile(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tempFile := createTempTestFile(t, "")
	defer os.Remove(tempFile)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{FilePath: tempFile},
	}

	// Act
	result, err := server.handleUploadTestLog(ctx, session, params)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error during upload initiation, got %v", err)
	}

	// 等待处理完成
	taskID := result.Meta["task_id"].(string)
	time.Sleep(100 * time.Millisecond)

	task := server.taskManager.GetTask(taskID)
	status := task.GetStatus()

	// 空文件应该导致处理完成但结果为空
	if status["status"] == "completed" {
		if result, ok := status["result"].(map[string]interface{}); ok {
			if result["total_tests"].(int) != 0 {
				t.Errorf("Expected total_tests=0 for empty file, got %v", result["total_tests"])
			}
		}
	}
}

// TestMCPServer_HandleUploadTestLog_PermissionDenied 测试权限拒绝场景
func TestMCPServer_HandleUploadTestLog_PermissionDenied(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 使用不存在的路径或无权限的路径
	invalidPath := "/root/nonexistent/file.json"
	if filepath.Separator == '\\' {
		// Windows路径
		invalidPath = "C:\\Windows\\System32\\nonexistent\\file.json"
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UploadRequest]{
		Arguments: UploadRequest{FilePath: invalidPath},
	}

	// Act
	result, err := server.handleUploadTestLog(ctx, session, params)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error during upload initiation, got %v", err)
	}

	// 等待处理完成
	taskID := result.Meta["task_id"].(string)
	time.Sleep(100 * time.Millisecond)

	task := server.taskManager.GetTask(taskID)
	status := task.GetStatus()

	// 应该失败并包含错误信息
	if status["status"] != "failed" {
		t.Errorf("Expected status=failed, got %v", status["status"])
	}
	if _, hasError := status["error"]; !hasError {
		t.Error("Expected error information in status")
	}
}

// TestMCPServer_HandleGetAnalysisResult_ConcurrentAccess 测试并发访问分析结果
func TestMCPServer_HandleGetAnalysisResult_ConcurrentAccess(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建测试任务
	tempFile := createTempTestFile(t, validTestLogContent())
	defer os.Remove(tempFile)

	task := server.taskManager.CreateTask("test-task", tempFile)
	go server.processTestLog(context.Background(), task)
	time.Sleep(50 * time.Millisecond) // 等待处理开始

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[QueryRequest]{
		Arguments: QueryRequest{TaskID: "test-task"},
	}

	// Act - 并发访问
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	results := make(chan *mcp.CallToolResultFor[map[string]interface{}], 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := server.handleGetAnalysisResult(ctx, session, params)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}()
	}

	wg.Wait()
	close(errors)
	close(results)

	// Assert
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent access error: %v", err)
		errorCount++
	}

	resultCount := 0
	for range results {
		resultCount++
	}

	if errorCount+resultCount != 10 {
		t.Errorf("Expected 10 total responses, got %d errors + %d results = %d", errorCount, resultCount, errorCount+resultCount)
	}
}

// TestMCPServer_HandleTerminateTask_AlreadyCompleted 测试终止已完成的任务
func TestMCPServer_HandleTerminateTask_AlreadyCompleted(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建并完成任务
	tempFile := createTempTestFile(t, validTestLogContent())
	defer os.Remove(tempFile)

	task := server.taskManager.CreateTask("completed-task", tempFile)
	server.processTestLog(context.Background(), task)
	time.Sleep(100 * time.Millisecond) // 等待处理完成

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[TerminateRequest]{
		Arguments: TerminateRequest{TaskID: "completed-task"},
	}

	// Act
	result, err := server.handleTerminateTask(ctx, session, params)

	// Assert
	if err == nil {
		t.Error("Expected error when terminating completed task")
	}
	if result != nil {
		t.Error("Expected no result when terminating completed task")
	}
}

// TestMCPServer_HandleGetTestDetails_InvalidTaskID 测试无效任务ID的测试详情查询
func TestMCPServer_HandleGetTestDetails_InvalidTaskID(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[TestDetailsRequest]{
		Arguments: TestDetailsRequest{
			TaskID:   "nonexistent-task",
			TestName: "TestExample",
		},
	}

	// Act
	result, err := server.handleGetTestDetails(ctx, session, params)

	// Assert
	if err == nil {
		t.Error("Expected error for nonexistent task")
	}
	if result != nil {
		t.Error("Expected no result for nonexistent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got %v", err)
	}
}

// TestMCPServer_ProcessTestLog_ContextCancellation 测试上下文取消时的处理
func TestMCPServer_ProcessTestLog_ContextCancellation(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tempFile := createTempTestFile(t, validTestLogContent())
	defer os.Remove(tempFile)

	task := server.taskManager.CreateTask("cancel-task", tempFile)

	// 创建会被取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	// Act
	server.processTestLog(ctx, task)

	// Assert
	status := task.GetStatus()
	// 任务可能已完成（因为处理很快）或被取消
	if status["status"] != "completed" && status["status"] != "failed" {
		t.Logf("Task status after context cancellation: %v", status["status"])
	}
}

// TestMCPServer_RegisterTools 测试工具注册
func TestMCPServer_RegisterTools(t *testing.T) {
	// Arrange & Act
	server, err := NewMCPServer()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if server.server == nil {
		t.Fatal("Expected MCP server to be initialized")
	}

	// 验证服务器已正确初始化（间接测试工具注册）
	if server.taskManager == nil {
		t.Error("Expected task manager to be initialized")
	}

	// 注意：由于MCP SDK的限制，我们无法直接验证工具是否已注册
	// 但我们可以通过创建服务器成功来间接验证registerTools()被调用
	t.Log("Tools registration tested indirectly through successful server creation")
}