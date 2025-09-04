package task

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/allanpk716/go_test_reader/internal/parser"
)

// TestNewManager 测试管理器创建
func TestNewManager(t *testing.T) {
	// Act
	manager := NewManager()

	// Assert
	if manager == nil {
		t.Fatal("Expected manager, got nil")
	}
	if manager.tasks == nil {
		t.Error("Expected tasks map to be initialized")
	}
}

// TestManager_CreateTask 测试任务创建
func TestManager_CreateTask(t *testing.T) {
	// Arrange
	manager := NewManager()
	filePath := "/test/path/test.json"

	// Act
	task := manager.CreateTask(filePath)

	// Assert
	if task == nil {
		t.Fatal("Expected task, got nil")
	}
	if task.ID == "" {
		t.Error("Expected task ID to be set")
	}
	if task.FilePath != filePath {
		t.Errorf("Expected FilePath=%s, got %s", filePath, task.FilePath)
	}
	if task.Status != StatusPending {
		t.Errorf("Expected Status=%s, got %s", StatusPending, task.Status)
	}
	if task.ctx == nil {
		t.Error("Expected context to be set")
	}
	if task.cancel == nil {
		t.Error("Expected cancel function to be set")
	}
}

// TestManager_GetTask 测试任务获取
func TestManager_GetTask(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")

	// Act
	retrievedTask := manager.GetTask(task.ID)

	// Assert
	if retrievedTask == nil {
		t.Fatal("Expected task, got nil")
	}
	if retrievedTask.ID != task.ID {
		t.Errorf("Expected ID=%s, got %s", task.ID, retrievedTask.ID)
	}
}

// TestManager_GetTask_NotFound 测试获取不存在的任务
func TestManager_GetTask_NotFound(t *testing.T) {
	// Arrange
	manager := NewManager()

	// Act
	task := manager.GetTask("non-existent-id")

	// Assert
	if task != nil {
		t.Errorf("Expected nil, got %v", task)
	}
}

// TestManager_TerminateTask 测试任务终止
func TestManager_TerminateTask(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	task.SetRunning()

	// Act
	err := manager.TerminateTask(task.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if task.Status != StatusCanceled {
		t.Errorf("Expected Status=%s, got %s", StatusCanceled, task.Status)
	}
	// 验证context是否被取消
	select {
	case <-task.ctx.Done():
		// 正确，context已被取消
	default:
		t.Error("Expected context to be canceled")
	}
}

// TestManager_TerminateTask_NotFound 测试终止不存在的任务
func TestManager_TerminateTask_NotFound(t *testing.T) {
	// Arrange
	manager := NewManager()

	// Act
	err := manager.TerminateTask("non-existent-id")

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "task not found: non-existent-id" {
		t.Errorf("Expected 'task not found' error, got %v", err)
	}
}

// TestManager_TerminateTask_AlreadyFinished 测试终止已完成的任务
func TestManager_TerminateTask_AlreadyFinished(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	task.SetResult(&parser.TestResult{TotalTests: 1})

	// Act
	err := manager.TerminateTask(task.ID)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !contains(err.Error(), "already finished") {
		t.Errorf("Expected 'already finished' error, got %v", err)
	}
}

// TestTask_SetRunning 测试设置任务为运行状态
func TestTask_SetRunning(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	initialTime := task.UpdatedAt

	// Act
	time.Sleep(1 * time.Millisecond) // 确保时间差异
	task.SetRunning()

	// Assert
	if task.Status != StatusRunning {
		t.Errorf("Expected Status=%s, got %s", StatusRunning, task.Status)
	}
	if !task.UpdatedAt.After(initialTime) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

// TestTask_SetResult 测试设置任务结果
func TestTask_SetResult(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	result := &parser.TestResult{
		TotalTests:  5,
		PassedTests: 3,
		FailedTests: 2,
	}

	// Act
	task.SetResult(result)

	// Assert
	if task.Status != StatusCompleted {
		t.Errorf("Expected Status=%s, got %s", StatusCompleted, task.Status)
	}
	if task.Result != result {
		t.Error("Expected result to be set")
	}
	if task.Result.TotalTests != 5 {
		t.Errorf("Expected TotalTests=5, got %d", task.Result.TotalTests)
	}
}

// TestTask_SetError 测试设置任务错误
func TestTask_SetError(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	testError := &testError{"test error"}

	// Act
	task.SetError(testError)

	// Assert
	if task.Status != StatusFailed {
		t.Errorf("Expected Status=%s, got %s", StatusFailed, task.Status)
	}
	if task.Error != testError {
		t.Error("Expected error to be set")
	}
}

// TestTask_GetStatus 测试获取任务状态
func TestTask_GetStatus(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	result := &parser.TestResult{
		TotalTests:      3,
		PassedTests:     2,
		FailedTests:     1,
		SkippedTests:    0,
		FailedTestNames: []string{"TestFailed"},
		PassedTestNames: []string{"TestPassed1", "TestPassed2"},
	}
	task.SetResult(result)

	// Act
	status := task.GetStatus()

	// Assert
	if status["task_id"] != task.ID {
		t.Errorf("Expected task_id=%s, got %v", task.ID, status["task_id"])
	}
	if status["status"] != string(StatusCompleted) {
		t.Errorf("Expected status=%s, got %v", StatusCompleted, status["status"])
	}
	if status["file_path"] != "/test/path" {
		t.Errorf("Expected file_path=/test/path, got %v", status["file_path"])
	}

	resultMap, ok := status["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}
	if resultMap["total_tests"] != 3 {
		t.Errorf("Expected total_tests=3, got %v", resultMap["total_tests"])
	}
}

// TestTask_GetTestDetails 测试获取测试详情
func TestTask_GetTestDetails(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	testDetails := map[string]*parser.TestDetail{
		"TestExample": {
			Status:  "pass",
			Output:  "test output",
			Error:   "",
			Elapsed: 0.001,
		},
	}
	result := &parser.TestResult{
		TestDetails: testDetails,
	}
	task.SetResult(result)

	// Act
	details := task.GetTestDetails("TestExample")

	// Assert
	if details == nil {
		t.Fatal("Expected details, got nil")
	}
	if details["test_name"] != "TestExample" {
		t.Errorf("Expected test_name=TestExample, got %v", details["test_name"])
	}
	if details["status"] != "pass" {
		t.Errorf("Expected status=pass, got %v", details["status"])
	}
}

// TestTask_GetTestDetails_NotFound 测试获取不存在的测试详情
func TestTask_GetTestDetails_NotFound(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	result := &parser.TestResult{
		TestDetails: make(map[string]*parser.TestDetail),
	}
	task.SetResult(result)

	// Act
	details := task.GetTestDetails("NonExistentTest")

	// Assert
	if details != nil {
		t.Errorf("Expected nil, got %v", details)
	}
}

// TestTask_IsCanceled 测试任务取消检查
func TestTask_IsCanceled(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")

	// Act & Assert - 初始状态
	if task.IsCanceled() {
		t.Error("Expected task not to be canceled initially")
	}

	// 取消任务
	task.cancel()

	// Act & Assert - 取消后
	if !task.IsCanceled() {
		t.Error("Expected task to be canceled after calling cancel")
	}
}

// TestManager_ConcurrentAccess 测试并发访问安全性
func TestManager_ConcurrentAccess(t *testing.T) {
	// Arrange
	manager := NewManager()
	const numGoroutines = 100
	var wg sync.WaitGroup
	taskIDs := make([]string, numGoroutines)

	// Act - 并发创建任务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			task := manager.CreateTask("/test/path")
			taskIDs[index] = task.ID
		}(i)
	}
	wg.Wait()

	// Assert - 验证所有任务都被正确创建
	for i, taskID := range taskIDs {
		if taskID == "" {
			t.Errorf("Task %d has empty ID", i)
		}
		task := manager.GetTask(taskID)
		if task == nil {
			t.Errorf("Task %d not found", i)
		}
	}

	// 验证ID唯一性
	idSet := make(map[string]bool)
	for _, taskID := range taskIDs {
		if idSet[taskID] {
			t.Errorf("Duplicate task ID found: %s", taskID)
		}
		idSet[taskID] = true
	}
}

// TestTask_ConcurrentStatusUpdates 测试并发状态更新
func TestTask_ConcurrentStatusUpdates(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")
	const numGoroutines = 50
	var wg sync.WaitGroup

	// Act - 并发更新任务状态
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if index%2 == 0 {
				task.SetRunning()
			} else {
				_ = task.GetStatus()
			}
		}(i)
	}
	wg.Wait()

	// Assert - 验证任务状态一致性
	status := task.GetStatus()
	if status == nil {
		t.Error("Expected status, got nil")
	}
}

// TestTask_ContextCancellation 测试上下文取消
func TestTask_ContextCancellation(t *testing.T) {
	// Arrange
	manager := NewManager()
	task := manager.CreateTask("/test/path")

	// 启动一个goroutine监听context取消
	done := make(chan bool)
	go func() {
		select {
		case <-task.ctx.Done():
			done <- true
		case <-time.After(5 * time.Second):
			done <- false
		}
	}()

	// Act
	task.cancel()

	// Assert
	select {
	case canceled := <-done:
		if !canceled {
			t.Error("Expected context to be canceled")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for context cancellation")
	}
}

// 辅助函数和类型
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}