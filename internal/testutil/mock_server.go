package testutil

import (
	"fmt"
	"sync"
	"time"

	"github.com/allanpk716/go_test_reader/internal/task"
)

// MockMCPServer Mock MCP服务器，用于集成测试
type MockMCPServer struct {
	mu           sync.RWMutex
	taskManager  *task.Manager
	uploadedLogs map[string]string // taskID -> log content
	responses    map[string]interface{} // method -> response
	callHistory  []MockCall
	isRunning    bool
}

// MockCall 记录方法调用历史
type MockCall struct {
	Method    string
	Params    map[string]interface{}
	Timestamp time.Time
	Result    interface{}
	Error     error
}

// NewMockMCPServer 创建Mock MCP服务器
func NewMockMCPServer() *MockMCPServer {
	return &MockMCPServer{
		taskManager:  task.NewManager(),
		uploadedLogs: make(map[string]string),
		responses:    make(map[string]interface{}),
		callHistory:  make([]MockCall, 0),
		isRunning:    false,
	}
}

// Start 启动Mock服务器
func (m *MockMCPServer) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true
}

// Stop 停止Mock服务器
func (m *MockMCPServer) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
}

// IsRunning 检查服务器是否运行中
func (m *MockMCPServer) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// SetResponse 设置方法的预期响应
func (m *MockMCPServer) SetResponse(method string, response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[method] = response
}

// GetCallHistory 获取调用历史
func (m *MockMCPServer) GetCallHistory() []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	history := make([]MockCall, len(m.callHistory))
	copy(history, m.callHistory)
	return history
}

// GetCallCount 获取指定方法的调用次数
func (m *MockMCPServer) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, call := range m.callHistory {
		if call.Method == method {
			count++
		}
	}
	return count
}

// ClearHistory 清空调用历史
func (m *MockMCPServer) ClearHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callHistory = m.callHistory[:0]
}

// recordCall 记录方法调用
func (m *MockMCPServer) recordCall(method string, params map[string]interface{}, result interface{}, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callHistory = append(m.callHistory, MockCall{
		Method:    method,
		Params:    params,
		Timestamp: time.Now(),
		Result:    result,
		Error:     err,
	})
}

// HandleUploadTestLog 模拟上传测试日志
func (m *MockMCPServer) HandleUploadTestLog(params map[string]interface{}) (interface{}, error) {
	if !m.IsRunning() {
		err := fmt.Errorf("server not running")
		m.recordCall("upload_test_log", params, nil, err)
		return nil, err
	}

	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		err := fmt.Errorf("invalid file_path parameter")
		m.recordCall("upload_test_log", params, nil, err)
		return nil, err
	}

	// 检查是否有预设响应
	if response, exists := m.responses["upload_test_log"]; exists {
		m.recordCall("upload_test_log", params, response, nil)
		return response, nil
	}

	// 默认行为：创建任务并开始处理
	taskID := fmt.Sprintf("mock-task-%d", time.Now().UnixNano())
	task := m.taskManager.CreateTask(taskID, filePath)
	if task == nil {
		err := fmt.Errorf("failed to create task")
		m.recordCall("upload_test_log", params, nil, err)
		return nil, err
	}

	// 模拟异步处理
	go m.simulateLogProcessing(taskID, filePath)

	result := map[string]interface{}{
		"task_id": taskID,
		"status":  "processing",
		"message": "Test log upload started",
	}

	m.recordCall("upload_test_log", params, result, nil)
	return result, nil
}

// HandleGetAnalysisResult 模拟获取分析结果
func (m *MockMCPServer) HandleGetAnalysisResult(params map[string]interface{}) (interface{}, error) {
	if !m.IsRunning() {
		err := fmt.Errorf("server not running")
		m.recordCall("get_analysis_result", params, nil, err)
		return nil, err
	}

	taskID, ok := params["task_id"].(string)
	if !ok || taskID == "" {
		err := fmt.Errorf("invalid task_id parameter")
		m.recordCall("get_analysis_result", params, nil, err)
		return nil, err
	}

	// 检查是否有预设响应
	if response, exists := m.responses["get_analysis_result"]; exists {
		m.recordCall("get_analysis_result", params, response, nil)
		return response, nil
	}

	// 默认行为：获取任务状态和结果
	task := m.taskManager.GetTask(taskID)
	if task == nil {
		err := fmt.Errorf("task not found: %s", taskID)
		m.recordCall("get_analysis_result", params, nil, err)
		return nil, err
	}

	result := task.GetStatus()

	m.recordCall("get_analysis_result", params, result, nil)
	return result, nil
}

// HandleTerminateTask 模拟终止任务
func (m *MockMCPServer) HandleTerminateTask(params map[string]interface{}) (interface{}, error) {
	if !m.IsRunning() {
		err := fmt.Errorf("server not running")
		m.recordCall("terminate_task", params, nil, err)
		return nil, err
	}

	taskID, ok := params["task_id"].(string)
	if !ok || taskID == "" {
		err := fmt.Errorf("invalid task_id parameter")
		m.recordCall("terminate_task", params, nil, err)
		return nil, err
	}

	// 检查是否有预设响应
	if response, exists := m.responses["terminate_task"]; exists {
		m.recordCall("terminate_task", params, response, nil)
		return response, nil
	}

	// 默认行为：终止任务
	err := m.taskManager.TerminateTask(taskID)
	if err != nil {
		m.recordCall("terminate_task", params, nil, err)
		return nil, err
	}

	result := map[string]interface{}{
		"task_id": taskID,
		"status":  "terminated",
		"message": "Task terminated successfully",
	}

	m.recordCall("terminate_task", params, result, nil)
	return result, nil
}

// HandleGetTestDetails 模拟获取测试详情
func (m *MockMCPServer) HandleGetTestDetails(params map[string]interface{}) (interface{}, error) {
	if !m.IsRunning() {
		err := fmt.Errorf("server not running")
		m.recordCall("get_test_details", params, nil, err)
		return nil, err
	}

	// 检查是否有预设响应
	if response, exists := m.responses["get_test_details"]; exists {
		m.recordCall("get_test_details", params, response, nil)
		return response, nil
	}

	// 默认行为：返回模拟测试详情
	mockResult := NewMockTestResult()
	details := mockResult.TestDetails

	m.recordCall("get_test_details", params, details, nil)
	return details, nil
}

// simulateLogProcessing 模拟日志处理过程
func (m *MockMCPServer) simulateLogProcessing(taskID, filePath string) {
	task := m.taskManager.GetTask(taskID)
	if task == nil {
		return
	}

	// 模拟处理延迟
	time.Sleep(100 * time.Millisecond)

	// 检查任务是否被取消
	if task.IsCanceled() {
		return
	}

	// 模拟解析结果
	mockResult := NewMockTestResult()
	parserResult := mockResult.ToParserTestResult()

	// 存储日志内容（模拟）
	m.mu.Lock()
	m.uploadedLogs[taskID] = fmt.Sprintf("mock log content for %s", filePath)
	m.mu.Unlock()

	// 设置任务结果
	task.SetResult(parserResult)
}

// GetUploadedLog 获取上传的日志内容
func (m *MockMCPServer) GetUploadedLog(taskID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	content, exists := m.uploadedLogs[taskID]
	return content, exists
}

// SimulateError 模拟错误情况
func (m *MockMCPServer) SimulateError(taskID string, err error) {
	task := m.taskManager.GetTask(taskID)
	if task != nil {
		task.SetError(err)
	}
}

// GetTaskManager 获取任务管理器（用于测试）
func (m *MockMCPServer) GetTaskManager() *task.Manager {
	return m.taskManager
}

// MockTestLogGenerator Mock测试日志生成器
type MockTestLogGenerator struct {
	scenarios map[string]string
}

// NewMockTestLogGenerator 创建Mock测试日志生成器
func NewMockTestLogGenerator() *MockTestLogGenerator {
	g := &MockTestLogGenerator{
		scenarios: make(map[string]string),
	}
	g.initializeScenarios()
	return g
}

// initializeScenarios 初始化测试场景
func (g *MockTestLogGenerator) initializeScenarios() {
	g.scenarios["success"] = `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestSuccess"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestSuccess","Elapsed":0.001}`

	g.scenarios["failure"] = `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestFailure"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestFailure","Output":"FAIL: assertion failed\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"fail","Package":"example/pkg","Test":"TestFailure","Elapsed":0.002}`

	g.scenarios["skip"] = `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestSkip"}
{"Time":"2024-01-15T10:30:01Z","Action":"skip","Package":"example/pkg","Test":"TestSkip","Elapsed":0.001}`

	g.scenarios["mixed"] = `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestPass"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestPass","Elapsed":0.001}
{"Time":"2024-01-15T10:30:02Z","Action":"run","Package":"example/pkg","Test":"TestFail"}
{"Time":"2024-01-15T10:30:03Z","Action":"fail","Package":"example/pkg","Test":"TestFail","Elapsed":0.002}
{"Time":"2024-01-15T10:30:04Z","Action":"run","Package":"example/pkg","Test":"TestSkip"}
{"Time":"2024-01-15T10:30:05Z","Action":"skip","Package":"example/pkg","Test":"TestSkip","Elapsed":0.001}`

	g.scenarios["invalid"] = `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestInvalid"}
invalid json line
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestInvalid","Elapsed":0.001}`
}

// GetScenario 获取指定场景的测试日志
func (g *MockTestLogGenerator) GetScenario(scenario string) (string, bool) {
	content, exists := g.scenarios[scenario]
	return content, exists
}

// GetAllScenarios 获取所有场景名称
func (g *MockTestLogGenerator) GetAllScenarios() []string {
	scenarios := make([]string, 0, len(g.scenarios))
	for name := range g.scenarios {
		scenarios = append(scenarios, name)
	}
	return scenarios
}

// AddScenario 添加自定义场景
func (g *MockTestLogGenerator) AddScenario(name, content string) {
	g.scenarios[name] = content
}