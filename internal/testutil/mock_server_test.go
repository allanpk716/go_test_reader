package testutil

import (
	"testing"
	"time"
)
// TestNewMockMCPServer 测试创建Mock MCP服务器
func TestNewMockMCPServer(t *testing.T) {
	// Act
	server := NewMockMCPServer()

	// Assert
	if server == nil {
		t.Fatal("Expected server, got nil")
	}
	if server.IsRunning() {
		t.Error("Expected server to be stopped initially")
	}
}

// TestMockMCPServer_StartStop 测试启动和停止服务器
func TestMockMCPServer_StartStop(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()

	// Act - Start
	server.Start()

	// Assert - Start
	if !server.IsRunning() {
		t.Error("Expected server to be running after start")
	}

	// Act - Stop
	server.Stop()

	// Assert - Stop
	if server.IsRunning() {
		t.Error("Expected server to be stopped after stop")
	}
}

// TestMockMCPServer_SetResponse 测试设置响应
func TestMockMCPServer_SetResponse(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()
	method := "upload_test_log"
	response := map[string]interface{}{"status": "success"}

	// Act
	server.SetResponse(method, response)

	// Assert - 无法直接验证内部状态，但确保方法不会panic
	// 这个测试主要验证方法调用不会出错
}

// TestMockMCPServer_GetCallHistory 测试获取调用历史
func TestMockMCPServer_GetCallHistory(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()

	// Act
	history := server.GetCallHistory()

	// Assert
	if history == nil {
		t.Error("Expected non-nil history")
	}
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(history))
	}
}

// TestMockMCPServer_GetCallCount 测试获取调用次数
func TestMockMCPServer_GetCallCount(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()

	// Act
	count := server.GetCallCount("upload_test_log")

	// Assert
	if count != 0 {
		t.Errorf("Expected initial call count=0, got %d", count)
	}

	// Test count for non-existent method
	count = server.GetCallCount("non_existent_method")
	if count != 0 {
		t.Errorf("Expected call count 0 for non-existent method, got %d", count)
	}

	// Test count for empty method name
	count = server.GetCallCount("")
	if count != 0 {
		t.Errorf("Expected call count 0 for empty method name, got %d", count)
	}
}

// TestMockMCPServer_ClearHistory 测试清空历史记录
func TestMockMCPServer_ClearHistory(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()

	// Act
	server.ClearHistory()

	// Assert
	history := server.GetCallHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d items", len(history))
	}
}

// TestNewMockTestLogGenerator 测试创建Mock测试日志生成器
func TestNewMockTestLogGenerator(t *testing.T) {
	// Act
	generator := NewMockTestLogGenerator()

	// Assert
	if generator == nil {
		t.Fatal("Expected generator, got nil")
	}
}

// TestMockTestLogGenerator_GetScenario 测试获取场景日志
func TestMockTestLogGenerator_GetScenario(t *testing.T) {
	// Arrange
	generator := NewMockTestLogGenerator()
	testCases := []struct {
		scenario string
		expected bool // 是否期望存在
	}{
		{"success", true},
		{"failure", true},
		{"skip", true},
		{"mixed", true},
		{"invalid", true},
		{"unknown", false}, // 未知场景不存在
	}

	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			// Act
			log, exists := generator.GetScenario(tc.scenario)

			// Assert
			if tc.expected && !exists {
				t.Errorf("Expected scenario %s to exist", tc.scenario)
			}
			if tc.expected && log == "" {
				t.Errorf("Expected non-empty log for scenario %s", tc.scenario)
			}
			if !tc.expected && exists {
				t.Errorf("Expected scenario %s to not exist", tc.scenario)
			}
		})
	}
}

// TestMockTestLogGenerator_GetAllScenarios 测试获取所有场景
func TestMockTestLogGenerator_GetAllScenarios(t *testing.T) {
	// Arrange
	generator := NewMockTestLogGenerator()

	// Act
	scenarios := generator.GetAllScenarios()

	// Assert
	if len(scenarios) == 0 {
		t.Error("Expected at least one scenario")
	}
	
	// 验证包含预期的场景
	expectedScenarios := []string{"success", "failure", "skip", "mixed", "invalid"}
	for _, expected := range expectedScenarios {
		found := false
		for _, actual := range scenarios {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find scenario %s", expected)
		}
	}
}

// TestMockTestLogGenerator_AddScenario 测试添加自定义场景
func TestMockTestLogGenerator_AddScenario(t *testing.T) {
	// Arrange
	generator := NewMockTestLogGenerator()
	customScenario := "custom"
	customContent := "custom test log content"

	// Act
	generator.AddScenario(customScenario, customContent)

	// Assert
	log, exists := generator.GetScenario(customScenario)
	if !exists {
		t.Errorf("Expected custom scenario to exist")
	}
	if log != customContent {
		t.Errorf("Expected log=%s, got %s", customContent, log)
	}
}

// TestMockCall_Structure 测试MockCall结构体
func TestMockCall_Structure(t *testing.T) {
	// Arrange & Act
	call := MockCall{
		Method:    "test_method",
		Params:    map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
		Result:    "test_result",
		Error:     nil,
	}

	// Assert
	if call.Method != "test_method" {
		t.Errorf("Expected Method=test_method, got %s", call.Method)
	}
	if call.Params["key"] != "value" {
		t.Errorf("Expected Params[key]=value, got %v", call.Params["key"])
	}
	if call.Result != "test_result" {
		t.Errorf("Expected Result=test_result, got %v", call.Result)
	}
	if call.Error != nil {
		t.Errorf("Expected Error=nil, got %v", call.Error)
	}
}

// TestMockMCPServer_HandleUploadTestLog 测试上传测试日志处理
func TestMockMCPServer_HandleUploadTestLog(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()
	server.Start()
	params := map[string]interface{}{
		"file_path": "/test/path/test.json",
	}

	// Act
	result, err := server.HandleUploadTestLog(params)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// 验证调用历史
	history := server.GetCallHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 call in history, got %d", len(history))
	}
	if history[0].Method != "upload_test_log" {
		t.Errorf("Expected method upload_test_log, got %s", history[0].Method)
	}
}

// TestMockMCPServer_HandleUploadTestLog_ErrorCases 测试上传测试日志错误处理
func TestMockMCPServer_HandleUploadTestLog_ErrorCases(t *testing.T) {
	server := NewMockMCPServer()
	server.Start()

	// Test with nil params
	result, err := server.HandleUploadTestLog(nil)
	if err == nil {
		t.Error("Expected error for nil params")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Test with empty params
	params := map[string]interface{}{}
	result, err = server.HandleUploadTestLog(params)
	if err == nil {
		t.Error("Expected error for empty params")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Test with invalid file_path type
	params = map[string]interface{}{
		"file_path": 123,
	}
	result, err = server.HandleUploadTestLog(params)
	if err == nil {
		t.Error("Expected error for invalid file_path type")
	}
}

// TestMockMCPServer_HandleGetAnalysisResult 测试获取分析结果处理
func TestMockMCPServer_HandleGetAnalysisResult(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()
	server.Start()
	taskID := "test-task-123"
	expectedResponse := map[string]interface{}{"status": "completed", "result": "test result"}
	server.SetResponse("get_analysis_result", expectedResponse)
	params := map[string]interface{}{
		"task_id": taskID,
	}

	// Act
	result, err := server.HandleGetAnalysisResult(params)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// 验证调用历史
	history := server.GetCallHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 call in history, got %d", len(history))
	}
	if history[0].Method != "get_analysis_result" {
		t.Errorf("Expected method get_analysis_result, got %s", history[0].Method)
	}
}

// TestMockMCPServer_HandleGetAnalysisResult_ErrorCases 测试获取分析结果错误处理
func TestMockMCPServer_HandleGetAnalysisResult_ErrorCases(t *testing.T) {
	server := NewMockMCPServer()
	server.Start()

	// Test with nil params
	result, err := server.HandleGetAnalysisResult(nil)
	if err == nil {
		t.Error("Expected error for nil params")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Test missing task_id
	params := map[string]interface{}{}
	result, err = server.HandleGetAnalysisResult(params)
	if err == nil {
		t.Error("Expected error for missing task_id")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Test invalid task_id type
	params = map[string]interface{}{
		"task_id": 123,
	}
	result, err = server.HandleGetAnalysisResult(params)
	if err == nil {
		t.Error("Expected error for invalid task_id type")
	}
}

// TestMockMCPServer_HandleTerminateTask 测试终止任务处理
func TestMockMCPServer_HandleTerminateTask(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()
	server.Start()
	taskID := "test-task-456"
	expectedResponse := map[string]interface{}{"status": "terminated"}
	server.SetResponse("terminate_task", expectedResponse)
	params := map[string]interface{}{
		"task_id": taskID,
	}

	// Act
	result, err := server.HandleTerminateTask(params)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// 验证调用历史
	history := server.GetCallHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 call in history, got %d", len(history))
	}
	if history[0].Method != "terminate_task" {
		t.Errorf("Expected method terminate_task, got %s", history[0].Method)
	}
}

// TestMockMCPServer_HandleTerminateTask_ErrorCases 测试终止任务错误处理
func TestMockMCPServer_HandleTerminateTask_ErrorCases(t *testing.T) {
	server := NewMockMCPServer()
	server.Start()

	// Test with nil params
	result, err := server.HandleTerminateTask(nil)
	if err == nil {
		t.Error("Expected error for nil params")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Test missing task_id
	params := map[string]interface{}{}
	result, err = server.HandleTerminateTask(params)
	if err == nil {
		t.Error("Expected error for missing task_id")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Test invalid task_id type
	params = map[string]interface{}{
		"task_id": 123,
	}
	result, err = server.HandleTerminateTask(params)
	if err == nil {
		t.Error("Expected error for invalid task_id type")
	}
}

// TestMockMCPServer_HandleGetTestDetails 测试获取测试详情处理
func TestMockMCPServer_HandleGetTestDetails(t *testing.T) {
	// Arrange
	server := NewMockMCPServer()
	server.Start()
	taskID := "test-task-789"
	expectedResponse := map[string]interface{}{"tests": []map[string]interface{}{{"name": "TestExample", "status": "pass"}}}
	server.SetResponse("get_test_details", expectedResponse)
	params := map[string]interface{}{
		"task_id": taskID,
	}

	// Act
	result, err := server.HandleGetTestDetails(params)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// 验证调用历史
	history := server.GetCallHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 call in history, got %d", len(history))
	}
	if history[0].Method != "get_test_details" {
		t.Errorf("Expected method get_test_details, got %s", history[0].Method)
	}
}

// TestMockMCPServer_HandleGetTestDetails_ErrorCases 测试获取测试详情错误处理
func TestMockMCPServer_HandleGetTestDetails_ErrorCases(t *testing.T) {
	server := NewMockMCPServer()

	// Test when server is not running
	result, err := server.HandleGetTestDetails(nil)
	if err == nil {
		t.Error("Expected error when server not running")
	}
	if result != nil {
		t.Error("Expected nil result for error case")
	}

	// Start server and test normal cases
	server.Start()

	// Test with nil params (should succeed)
	result, err = server.HandleGetTestDetails(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Test with empty params (should succeed)
	params := map[string]interface{}{}
	result, err = server.HandleGetTestDetails(params)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}