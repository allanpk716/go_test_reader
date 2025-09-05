package server

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/allanpk716/go_test_reader/internal/parser"
)

// MCPServer MCP 服务器实例
type MCPServer struct {
	server *mcp.Server
}

// AnalyzeTestLogRequest 分析测试日志请求参数
type AnalyzeTestLogRequest struct {
	FilePath string `json:"file_path"`
}

// TestOverviewResponse 测试总览响应
type TestOverviewResponse struct {
	AllTestsPassed   bool     `json:"all_tests_passed"`
	TotalTests       int      `json:"total_tests"`
	FailedTestsCount int      `json:"failed_tests_count"`
	FailedTestNames  []string `json:"failed_test_names"`
}

// GetTestDetailsRequest 获取测试详情请求参数
type GetTestDetailsRequest struct {
	FilePath string `json:"file_path"`
	TestName string `json:"test_name"`
}

// TestDetailsResponse 测试详情响应
type TestDetailsResponse struct {
	TestName string `json:"test_name"`
	Status   string `json:"status"`
	Output   string `json:"output"`
	Error    string `json:"error"`
	Elapsed  float64 `json:"elapsed"`
}

// NewMCPServer 创建新的 MCP 服务器
func NewMCPServer() (*MCPServer, error) {
	// 创建 MCP 服务器
	server := mcp.NewServer("go-test-reader", "1.0.0", nil)
	
	mcpServer := &MCPServer{
		server: server,
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
	// 注册测试日志分析工具
	analyzeTool := mcp.NewServerTool(
		"analyze_test_log",
		"分析 go test 输出的测试日志文件，返回测试总览信息",
		s.handleAnalyzeTestLog,
	)
	
	// 注册测试详情查询工具
	detailsTool := mcp.NewServerTool(
		"get_test_details",
		"根据文件路径和测试名称获取详细错误信息",
		s.handleGetTestDetails,
	)
	
	// 添加工具到服务器
	s.server.AddTools(analyzeTool, detailsTool)
}

// handleAnalyzeTestLog 处理测试日志分析
func (s *MCPServer) handleAnalyzeTestLog(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[AnalyzeTestLogRequest]) (*mcp.CallToolResultFor[TestOverviewResponse], error) {
	filePath := params.Arguments.FilePath
	if filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}
	
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// 解析测试日志
	result, err := s.parseTestLogWithAutoDetection(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test log: %w", err)
	}
	
	// 构建响应
	allTestsPassed := result.FailedTests == 0
	response := TestOverviewResponse{
		AllTestsPassed:   allTestsPassed,
		TotalTests:       result.TotalTests,
		FailedTestsCount: result.FailedTests,
		FailedTestNames:  result.FailedTestNames,
	}
	
	return &mcp.CallToolResultFor[TestOverviewResponse]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("测试分析完成：总计 %d 个测试，%d 个失败", result.TotalTests, result.FailedTests),
			},
		},
		Meta: mcp.Meta{
			"all_tests_passed":   response.AllTestsPassed,
			"total_tests":        response.TotalTests,
			"failed_tests_count": response.FailedTestsCount,
			"failed_test_names":  response.FailedTestNames,
		},
	}, nil
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



// handleGetTestDetails 获取测试详情
func (s *MCPServer) handleGetTestDetails(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetTestDetailsRequest]) (*mcp.CallToolResultFor[TestDetailsResponse], error) {
	filePath := params.Arguments.FilePath
	if filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}
	
	testName := params.Arguments.TestName
	if testName == "" {
		return nil, fmt.Errorf("test_name parameter is required")
	}
	
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// 解析测试日志
	result, err := s.parseTestLogWithAutoDetection(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test log: %w", err)
	}
	
	// 查找指定的测试详情
	testDetail, exists := result.TestDetails[testName]
	if !exists {
		return nil, fmt.Errorf("test not found: %s", testName)
	}
	
	// 构建响应
	response := TestDetailsResponse{
		TestName: testName,
		Status:   testDetail.Status,
		Output:   testDetail.Output,
		Error:    testDetail.Error,
		Elapsed:  testDetail.Elapsed,
	}
	
	return &mcp.CallToolResultFor[TestDetailsResponse]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("测试 %s 的详细信息", testName),
			},
		},
		Meta: mcp.Meta{
			"test_name": response.TestName,
			"status":    response.Status,
			"output":    response.Output,
			"error":     response.Error,
			"elapsed":   response.Elapsed,
		},
	}, nil
}