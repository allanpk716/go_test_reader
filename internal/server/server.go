package server

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/allanpk716/go_test_reader/internal/parser"
	"github.com/allanpk716/go_test_reader/internal/task"
)

// MCPServer MCP 服务器实例
type MCPServer struct {
	server      *mcp.Server
	taskManager *task.Manager
	mu          sync.RWMutex
}

// UploadRequest 上传请求参数
type UploadRequest struct {
	FilePath string `json:"file_path"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// QueryRequest 查询请求参数
type QueryRequest struct {
	TaskID string `json:"task_id"`
}

// TerminateRequest 终止请求参数
type TerminateRequest struct {
	TaskID string `json:"task_id"`
}

// TerminateResponse 终止响应
type TerminateResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TestDetailsRequest 测试详情请求参数
type TestDetailsRequest struct {
	TaskID   string `json:"task_id"`
	TestName string `json:"test_name"`
}

// NewMCPServer 创建新的 MCP 服务器
func NewMCPServer() (*MCPServer, error) {
	// 创建任务管理器
	taskManager := task.NewManager()
	
	// 创建 MCP 服务器
	server := mcp.NewServer("go-test-reader", "1.0.0", nil)
	
	mcpServer := &MCPServer{
		server:      server,
		taskManager: taskManager,
	}
	
	// 注册工具
	mcpServer.registerTools()
	
	return mcpServer, nil
}

// Run 启动服务器
func (s *MCPServer) Run(ctx context.Context) error {
	transport := mcp.NewStdioTransport()
	return s.server.Run(ctx, transport)
}

// registerTools 注册 MCP 工具
func (s *MCPServer) registerTools() {
	// 注册文件上传工具
	uploadTool := mcp.NewServerTool(
		"upload_test_log",
		"上传 go test -json 输出的测试日志文件进行分析",
		s.handleUploadTestLog,
	)
	
	// 注册查询结果工具
	queryTool := mcp.NewServerTool(
		"get_analysis_result",
		"根据任务ID获取分析结果",
		s.handleGetAnalysisResult,
	)
	
	// 注册任务终止工具
	terminateTool := mcp.NewServerTool(
		"terminate_task",
		"终止指定的分析任务",
		s.handleTerminateTask,
	)
	
	// 注册测试详情查询工具
	detailsTool := mcp.NewServerTool(
		"get_test_details",
		"根据任务ID和测试名称获取详细错误信息",
		s.handleGetTestDetails,
	)
	
	// 添加工具到服务器
	s.server.AddTools(uploadTool, queryTool, terminateTool, detailsTool)
}

// handleUploadTestLog 处理文件上传
func (s *MCPServer) handleUploadTestLog(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadRequest]) (*mcp.CallToolResultFor[UploadResponse], error) {
	filePath := params.Arguments.FilePath
	if filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}
	
	// 生成任务ID
	taskID := uuid.New().String()
	
	// 创建并启动分析任务
	task := s.taskManager.CreateTask(taskID, filePath)
	go s.processTestLog(ctx, task)
	
	response := UploadResponse{
		TaskID:  taskID,
		Status:  "started",
		Message: "测试日志分析任务已启动",
	}
	
	return &mcp.CallToolResultFor[UploadResponse]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("任务已创建，ID: %s", taskID),
			},
		},
		Meta: mcp.Meta{
			"task_id": response.TaskID,
			"status":  response.Status,
			"message": response.Message,
		},
	}, nil
}

// processTestLog 处理测试日志文件
func (s *MCPServer) processTestLog(ctx context.Context, task *task.Task) {
	defer func() {
		if r := recover(); r != nil {
			task.SetError(fmt.Errorf("panic: %v", r))
		}
	}()
	
	// 打开文件
	file, err := os.Open(task.FilePath)
	if err != nil {
		task.SetError(fmt.Errorf("failed to open file: %w", err))
		return
	}
	defer file.Close()
	
	// 自动检测文件格式并选择合适的解析器
	result, err := s.parseTestLogWithAutoDetection(file)
	if err != nil {
		task.SetError(fmt.Errorf("failed to parse test log: %w", err))
		return
	}
	
	// 设置结果
	task.SetResult(result)
}

// parseTestLogWithAutoDetection 自动检测文件格式并解析
func (s *MCPServer) parseTestLogWithAutoDetection(file *os.File) (*parser.TestResult, error) {
	// 首先尝试检测是否为 JSON 格式
	file.Seek(0, 0) // 重置文件指针
	if err := parser.ValidateTestLog(file); err == nil {
		// 是 JSON 格式，使用 JSON 解析器
		file.Seek(0, 0) // 重置文件指针
		return parser.ParseTestLog(file)
	}
	
	// 尝试检测是否为文本格式
	file.Seek(0, 0) // 重置文件指针
	if err := parser.ValidateTestTextLog(file); err == nil {
		// 是文本格式，使用文本解析器
		file.Seek(0, 0) // 重置文件指针
		return parser.ParseTestTextLog(file)
	}
	
	// 如果两种格式都不匹配，返回错误
	return nil, fmt.Errorf("file does not appear to be valid go test output (neither JSON nor text format)")
}

// handleGetAnalysisResult 获取分析结果
func (s *MCPServer) handleGetAnalysisResult(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[QueryRequest]) (*mcp.CallToolResultFor[map[string]interface{}], error) {
	taskID := params.Arguments.TaskID
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}
	
	task := s.taskManager.GetTask(taskID)
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	status := task.GetStatus()
	
	return &mcp.CallToolResultFor[map[string]interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("任务 %s 状态: %s", taskID, status["status"]),
			},
		},
		Meta: mcp.Meta(status),
	}, nil
}

// handleTerminateTask 终止任务
func (s *MCPServer) handleTerminateTask(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[TerminateRequest]) (*mcp.CallToolResultFor[TerminateResponse], error) {
	taskID := params.Arguments.TaskID
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}
	
	err := s.taskManager.TerminateTask(taskID)
	if err != nil {
		return nil, err
	}
	
	response := TerminateResponse{
		TaskID:  taskID,
		Status:  "terminated",
		Message: "任务已终止",
	}
	
	return &mcp.CallToolResultFor[TerminateResponse]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("任务 %s 已终止", taskID),
			},
		},
		Meta: mcp.Meta{
			"task_id": response.TaskID,
			"status":  response.Status,
			"message": response.Message,
		},
	}, nil
}

// handleGetTestDetails 获取测试详情
func (s *MCPServer) handleGetTestDetails(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[TestDetailsRequest]) (*mcp.CallToolResultFor[map[string]interface{}], error) {
	taskID := params.Arguments.TaskID
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}
	
	testName := params.Arguments.TestName
	if testName == "" {
		return nil, fmt.Errorf("test_name parameter is required")
	}
	
	task := s.taskManager.GetTask(taskID)
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	details := task.GetTestDetails(testName)
	if details == nil {
		return nil, fmt.Errorf("test not found: %s", testName)
	}
	
	return &mcp.CallToolResultFor[map[string]interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("测试 %s 的详细信息", testName),
			},
		},
		Meta: mcp.Meta(details),
	}, nil
}