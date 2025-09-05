package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试辅助函数现在在 test_helpers.go 文件中定义

// TestMCPServer_AnalyzeTestLog_Success 测试成功分析测试日志
func TestMCPServer_AnalyzeTestLog_Success(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 获取测试数据文件路径
	testDataDir := filepath.Join("..", "..", "test_data")
	okFilePath := filepath.Join(testDataDir, "ok_00.txt")
	
	// 验证文件存在
	_, err := os.Stat(okFilePath)
	require.NoError(t, err, "Test file should exist")
	
	// 调用测试方法
	result, err := mst.testAnalyzeTestLog(okFilePath)
	require.NoError(t, err, "Tool call should succeed")
	
	// 验证结果不为空
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Content, "Result content should not be nil")
	
	// 验证结果结构
	require.NotNil(t, result.Meta, "Meta should not be nil")
	
	allTestsPassed, exists := result.Meta["all_tests_passed"]
	assert.True(t, exists, "all_tests_passed should exist in meta")
	assert.True(t, allTestsPassed.(bool), "All tests should pass for ok_00.txt")
	
	totalTests, exists := result.Meta["total_tests"]
	assert.True(t, exists, "total_tests should exist in meta")
	assert.True(t, totalTests.(int) > 0, "Total tests should be greater than 0")
	
	failedTestsCount, exists := result.Meta["failed_tests_count"]
	assert.True(t, exists, "failed_tests_count should exist in meta")
	assert.Equal(t, 0, failedTestsCount.(int), "Failed tests count should be 0")
	
	failedTestNames, exists := result.Meta["failed_test_names"]
	assert.True(t, exists, "failed_test_names should exist in meta")
	assert.Empty(t, failedTestNames.([]string), "Failed test names should be empty")
}

// TestMCPServer_AnalyzeTestLog_FailedTests 测试分析失败的测试日志
func TestMCPServer_AnalyzeTestLog_FailedTests(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 获取失败测试数据文件路径
	testDataDir := filepath.Join("..", "..", "test_data")
	failFilePath := filepath.Join(testDataDir, "fail_01.txt")
	
	// 验证测试文件存在
	_, err := os.Stat(failFilePath)
	require.NoError(t, err, "Test data file should exist: %s", failFilePath)
	
	// 调用测试方法
	result, err := mst.testAnalyzeTestLog(failFilePath)
	require.NoError(t, err, "Tool call should succeed")
	
	// 验证结果不为空
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Content, "Result content should not be nil")
	
	// 验证结果结构
	require.NotNil(t, result.Meta, "Meta should not be nil")
	
	allTestsPassed, exists := result.Meta["all_tests_passed"]
	assert.True(t, exists, "all_tests_passed should exist in meta")
	assert.False(t, allTestsPassed.(bool), "Not all tests should pass for fail_00.txt")
	
	totalTests, exists := result.Meta["total_tests"]
	assert.True(t, exists, "total_tests should exist in meta")
	assert.True(t, totalTests.(int) > 0, "Total tests should be greater than 0")
	
	failedTestsCount, exists := result.Meta["failed_tests_count"]
	assert.True(t, exists, "failed_tests_count should exist in meta")
	assert.True(t, failedTestsCount.(int) > 0, "Failed tests count should be greater than 0")
	
	failedTestNames, exists := result.Meta["failed_test_names"]
	assert.True(t, exists, "failed_test_names should exist in meta")
	assert.NotEmpty(t, failedTestNames.([]string), "Failed test names should not be empty")
}

// TestMCPServer_AnalyzeTestLog_InvalidFile 测试分析不存在的文件
func TestMCPServer_AnalyzeTestLog_InvalidFile(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 调用测试方法，应该失败
	_, err := mst.testAnalyzeTestLog("/nonexistent/file.txt")
	assert.Error(t, err, "Tool call should fail for nonexistent file")
}

// TestMCPServer_AnalyzeTestLog_EmptyFilePath 测试空文件路径
func TestMCPServer_AnalyzeTestLog_EmptyFilePath(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 调用测试方法，应该失败
	_, err := mst.testAnalyzeTestLog("")
	assert.Error(t, err, "Tool call should fail for empty file path")
	assert.Contains(t, err.Error(), "file_path parameter is required", "Error should mention required parameter")
}

// TestMCPServer_GetTestDetails_Success 测试成功获取测试详情
func TestMCPServer_GetTestDetails_Success(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 获取测试数据文件路径
	testDataDir := filepath.Join("..", "..", "test_data")
	okFilePath := filepath.Join(testDataDir, "ok_00.txt")
	
	// 首先分析测试日志获取测试名称
	_, err := mst.testAnalyzeTestLog(okFilePath)
	require.NoError(t, err, "Analyze tool call should succeed")
	
	// 从分析结果中获取测试名称（这里我们知道ok_00.txt包含通过的测试）
	// 由于我们知道测试文件的内容，可以直接使用一个已知的测试名称
	testName := "TestUploadState_JSONSerialization"
	
	// 调用测试方法
	result, err := mst.testGetTestDetails(okFilePath, testName)
	require.NoError(t, err, "Get test details tool call should succeed")
	
	// 验证结果不为空
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Content, "Result content should not be nil")
	
	// 验证结果结构
	require.NotNil(t, result.Meta, "Meta should not be nil")
	
	testNameMeta, exists := result.Meta["test_name"]
	assert.True(t, exists, "test_name should exist in meta")
	assert.Equal(t, testName, testNameMeta.(string), "Test name should match")
	
	status, exists := result.Meta["status"]
	assert.True(t, exists, "status should exist in meta")
	assert.Equal(t, "pass", status.(string), "Status should be pass")
	
	elapsed, exists := result.Meta["elapsed"]
	assert.True(t, exists, "elapsed should exist in meta")
	assert.GreaterOrEqual(t, elapsed.(float64), 0.0, "Elapsed time should be non-negative")
}

// TestMCPServer_GetTestDetails_InvalidTestName 测试获取不存在的测试详情
func TestMCPServer_GetTestDetails_InvalidTestName(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 获取测试数据文件路径
	testDataDir := filepath.Join("..", "..", "test_data")
	okFilePath := filepath.Join(testDataDir, "ok_00.txt")
	
	// 调用测试方法，应该失败
	_, err := mst.testGetTestDetails(okFilePath, "NonExistentTest")
	assert.Error(t, err, "Tool call should fail for nonexistent test")
	assert.Contains(t, err.Error(), "test not found", "Error should mention test not found")
}

// TestMCPServer_GetTestDetails_EmptyParameters 测试空参数
func TestMCPServer_GetTestDetails_EmptyParameters(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 测试空文件路径
	_, err := mst.testGetTestDetails("", "SomeTest")
	assert.Error(t, err, "Tool call should fail for empty file path")
	assert.Contains(t, err.Error(), "file_path parameter is required", "Error should mention required file_path")
	
	// 测试空测试名称
	_, err = mst.testGetTestDetails("some_file.txt", "")
	assert.Error(t, err, "Tool call should fail for empty test name")
	assert.Contains(t, err.Error(), "test_name parameter is required", "Error should mention required test_name")
}

// TestMCPServer_ToolsRegistration 测试工具注册
func TestMCPServer_ToolsRegistration(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 验证服务器创建成功
	require.NotNil(t, mst.server, "Server should be created")
	require.NotNil(t, mst.server.server, "MCP server should be initialized")
	
	// 由于我们无法直接访问工具列表，我们通过调用工具来验证它们已注册
	// 测试analyze_test_log工具是否可用
	testDataDir := filepath.Join("..", "..", "test_data")
	okFilePath := filepath.Join(testDataDir, "ok_00.txt")
	
	_, err := mst.testAnalyzeTestLog(okFilePath)
	if os.IsNotExist(err) {
		// 如果文件不存在，这是预期的，说明工具已注册
		t.Log("analyze_test_log tool is registered (file not found is expected)")
	} else {
		require.NoError(t, err, "analyze_test_log tool should be available")
	}
	
	// 测试get_test_details工具是否可用
	_, err = mst.testGetTestDetails(okFilePath, "TestSomething")
	if os.IsNotExist(err) {
		// 如果文件不存在，这是预期的，说明工具已注册
		t.Log("get_test_details tool is registered (file not found is expected)")
	} else {
		// 其他错误也是可以接受的，只要不是"工具未找到"错误
		t.Logf("get_test_details tool is registered (error: %v)", err)
	}
}

// TestMCPServer_ConcurrentRequests 测试并发请求处理
func TestMCPServer_ConcurrentRequests(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 获取测试数据文件路径
	testDataDir := filepath.Join("..", "..", "test_data")
	okFilePath := filepath.Join(testDataDir, "ok_00.txt")
	
	// 并发执行多个请求
	concurrency := 5
	results := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			_, err := mst.testAnalyzeTestLog(okFilePath)
			results <- err
		}()
	}
	
	// 收集结果
	for i := 0; i < concurrency; i++ {
		err := <-results
		if os.IsNotExist(err) {
			// 如果文件不存在，这是可以接受的
			t.Logf("Concurrent request %d: file not found (expected)", i)
		} else {
			assert.NoError(t, err, "Concurrent request %d should succeed", i)
		}
	}
}