package server

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// MCPServerTest 测试辅助结构体
type MCPServerTest struct {
	server *MCPServer
	ctx    context.Context
	cancel context.CancelFunc
}

// setupMCPServerTest 设置MCP服务器测试环境
func setupMCPServerTest(t *testing.T) *MCPServerTest {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	
	// 创建MCP服务器
	server, err := NewMCPServer()
	require.NoError(t, err, "Should create MCP server")
	
	return &MCPServerTest{
		server: server,
		ctx:    ctx,
		cancel: cancel,
	}
}

// teardownMCPServerTest 清理MCP服务器测试环境
func (mst *MCPServerTest) teardownMCPServerTest() {
	if mst.cancel != nil {
		mst.cancel()
	}
}

// testAnalyzeTestLog 辅助方法，用于测试分析测试日志功能
func (mst *MCPServerTest) testAnalyzeTestLog(filePath string) (*mcp.CallToolResultFor[TestOverviewResponse], error) {
	request := &AnalyzeTestLogRequest{
		FilePath: filePath,
	}
	
	params := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: *request,
	}
	
	return mst.server.handleAnalyzeTestLog(mst.ctx, nil, params)
}

// testAnalyzeTestLogWithContext 辅助方法，支持自定义上下文
func (mst *MCPServerTest) testAnalyzeTestLogWithContext(ctx context.Context, filePath string) (*mcp.CallToolResultFor[TestOverviewResponse], error) {
	request := &AnalyzeTestLogRequest{
		FilePath: filePath,
	}
	
	params := &mcp.CallToolParamsFor[AnalyzeTestLogRequest]{
		Arguments: *request,
	}
	
	return mst.server.handleAnalyzeTestLog(ctx, nil, params)
}

// testGetTestDetails 辅助方法，用于测试获取测试详情功能
func (mst *MCPServerTest) testGetTestDetails(filePath, testName string) (*mcp.CallToolResultFor[TestDetailsResponse], error) {
	request := &GetTestDetailsRequest{
		FilePath: filePath,
		TestName: testName,
	}
	
	params := &mcp.CallToolParamsFor[GetTestDetailsRequest]{
		Arguments: *request,
	}
	
	return mst.server.handleGetTestDetails(mst.ctx, nil, params)
}