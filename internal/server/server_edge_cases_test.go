package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServer_EdgeCases_LargeFiles 测试大文件处理
func TestMCPServer_EdgeCases_LargeFiles(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建一个大的测试文件
	tempFile, err := os.CreateTemp("", "large_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入大量测试数据
	largeContent := strings.Repeat("=== RUN   TestExample\n--- PASS: TestExample (0.00s)\n", 1000)
	largeContent += "PASS\nok      example.com/test      1.234s\n"
	
	_, err = tempFile.WriteString(largeContent)
	require.NoError(t, err, "Should write to temp file")
	tempFile.Close()
	
	// 测试分析大文件
	result, err := mst.testAnalyzeTestLog(tempFile.Name())
	assert.NoError(t, err, "Should handle large file successfully")
	assert.NotNil(t, result, "Result should not be nil")
}

// TestMCPServer_EdgeCases_EmptyFile 测试空文件处理
func TestMCPServer_EdgeCases_EmptyFile(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建空文件
	tempFile, err := os.CreateTemp("", "empty_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// 测试分析空文件
	_, err = mst.testAnalyzeTestLog(tempFile.Name())
	assert.Error(t, err, "Should fail to parse empty file")
}

// TestMCPServer_EdgeCases_BinaryFile 测试二进制文件处理
func TestMCPServer_EdgeCases_BinaryFile(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建二进制文件
	tempFile, err := os.CreateTemp("", "binary_test_*.bin")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入大量二进制数据，确保没有任何有效的JSON或测试模式
	binaryData := make([]byte, 1000)
	for i := range binaryData {
		binaryData[i] = byte(i % 256)
	}
	_, err = tempFile.Write(binaryData)
	require.NoError(t, err, "Should write binary data")
	tempFile.Close()
	
	// 测试分析二进制文件
	_, err = mst.testAnalyzeTestLog(tempFile.Name())
	assert.Error(t, err, "Should fail to parse binary file")
}

// TestMCPServer_EdgeCases_SpecialCharacters 测试特殊字符处理
func TestMCPServer_EdgeCases_SpecialCharacters(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建包含特殊字符的测试文件
	tempFile, err := os.CreateTemp("", "special_chars_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入包含特殊字符的测试数据
	specialContent := `=== RUN   Test中文测试
--- PASS: Test中文测试 (0.00s)
=== RUN   TestEmoji🚀
--- PASS: TestEmoji🚀 (0.00s)
=== RUN   TestSpecial!@#$%^&*()
--- PASS: TestSpecial!@#$%^&*() (0.00s)
PASS
ok      example.com/test      0.123s
`
	
	_, err = tempFile.WriteString(specialContent)
	require.NoError(t, err, "Should write special content")
	tempFile.Close()
	
	// 测试分析包含特殊字符的文件
	result, err := mst.testAnalyzeTestLog(tempFile.Name())
	assert.NoError(t, err, "Should handle special characters successfully")
	assert.NotNil(t, result, "Result should not be nil")
}

// TestMCPServer_EdgeCases_VeryLongTestNames 测试超长测试名称
func TestMCPServer_EdgeCases_VeryLongTestNames(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建包含超长测试名称的文件
	tempFile, err := os.CreateTemp("", "long_names_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 生成超长测试名称
	longTestName := "Test" + strings.Repeat("VeryLongTestName", 50)
	longContent := fmt.Sprintf("=== RUN   %s\n--- PASS: %s (0.00s)\nPASS\nok      example.com/test      0.123s\n", longTestName, longTestName)
	
	_, err = tempFile.WriteString(longContent)
	require.NoError(t, err, "Should write long content")
	tempFile.Close()
	
	// 测试分析包含超长测试名称的文件
	result, err := mst.testAnalyzeTestLog(tempFile.Name())
	assert.NoError(t, err, "Should handle long test names successfully")
	assert.NotNil(t, result, "Result should not be nil")
}

// TestMCPServer_EdgeCases_PermissionDenied 测试权限拒绝情况
func TestMCPServer_EdgeCases_PermissionDenied(t *testing.T) {
	// 跳过Windows上的权限测试，因为Windows权限模型不同
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建文件并设置为无读权限
	tempFile, err := os.CreateTemp("", "no_permission_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	
	_, err = tempFile.WriteString("=== RUN   TestExample\n--- PASS: TestExample (0.00s)\nPASS\n")
	require.NoError(t, err, "Should write to temp file")
	tempFile.Close()
	
	// 移除读权限
	err = os.Chmod(tempFile.Name(), 0000)
	require.NoError(t, err, "Should change file permissions")
	
	// 恢复权限以便清理
	defer os.Chmod(tempFile.Name(), 0644)
	
	// 测试访问无权限文件
	_, err = mst.testAnalyzeTestLog(tempFile.Name())
	assert.Error(t, err, "Should fail to access file without permission")
}

// TestMCPServer_EdgeCases_ConcurrentFileAccess 测试并发文件访问
func TestMCPServer_EdgeCases_ConcurrentFileAccess(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 获取测试数据文件
	testDataDir := filepath.Join("..", "..", "test_data")
	okFilePath := filepath.Join(testDataDir, "ok_00.txt")
	
	// 验证文件存在
	_, err := os.Stat(okFilePath)
	require.NoError(t, err, "Test file should exist")
	
	// 并发访问同一文件
	concurrency := 10
	results := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			_, err := mst.testAnalyzeTestLog(okFilePath)
			results <- err
		}(i)
	}
	
	// 收集所有结果
	for i := 0; i < concurrency; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent access %d should succeed", i)
	}
}

// TestMCPServer_EdgeCases_TimeoutHandling 测试超时处理
func TestMCPServer_EdgeCases_TimeoutHandling(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// 获取测试数据文件
	okFilePath := filepath.Join("..", "..", "test_data", "ok_00.txt")
	
	// 测试在超时上下文中分析文件
	// 由于文件很小，这应该在超时前完成
	_, err := mst.testAnalyzeTestLogWithContext(ctx, okFilePath)
	if err != nil {
		// 如果超时，这是预期的
		assert.Contains(t, err.Error(), "timeout", "Should timeout or complete successfully")
	}
}

// TestMCPServer_EdgeCases_MalformedJSON 测试格式错误的JSON
func TestMCPServer_EdgeCases_MalformedJSON(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建包含格式错误JSON的文件
	tempFile, err := os.CreateTemp("", "malformed_json_*.log")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入完全无效的内容，确保没有任何有效JSON行
	malformedContent := `This is not JSON at all
Another line of invalid content
{broken json without closing brace
[invalid array without closing bracket
"incomplete string
random text here
more invalid content
{"incomplete": "json"
["another", "broken", "array"
"just a string without context"`
	_, err = tempFile.WriteString(malformedContent)
	require.NoError(t, err, "Should write malformed content")
	tempFile.Close()
	
	// 测试分析格式错误的JSON文件
	_, err = mst.testAnalyzeTestLog(tempFile.Name())
	assert.Error(t, err, "Should fail to parse malformed JSON")
}

// TestMCPServer_EdgeCases_MixedFormat 测试混合格式文件
func TestMCPServer_EdgeCases_MixedFormat(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建包含混合格式的文件
	tempFile, err := os.CreateTemp("", "mixed_format_*.log")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入既不是有效JSON也不是有效测试文本的混合内容
	mixedContent := `This is some random text
{"incomplete": "json"
Not a test pattern here
["broken", "array"
Some more random content
{"another": "incomplete"
No test markers like RUN or PASS
Just random text content
More invalid JSON: {"key":
And some other content`
	_, err = tempFile.WriteString(mixedContent)
	require.NoError(t, err, "Should write mixed content")
	tempFile.Close()
	
	// 测试分析混合格式文件
	_, err = mst.testAnalyzeTestLog(tempFile.Name())
	assert.Error(t, err, "Should fail to parse mixed format file")
}

// TestMCPServer_EdgeCases_UnicodeHandling 测试Unicode处理
func TestMCPServer_EdgeCases_UnicodeHandling(t *testing.T) {
	mst := setupMCPServerTest(t)
	defer mst.teardownMCPServerTest()
	
	// 创建包含各种Unicode字符的文件
	tempFile, err := os.CreateTemp("", "unicode_test_*.txt")
	require.NoError(t, err, "Should create temp file")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入包含Unicode字符的测试数据
	unicodeContent := `=== RUN   Test_العربية
--- PASS: Test_العربية (0.00s)
=== RUN   Test_русский
--- PASS: Test_русский (0.00s)
=== RUN   Test_日本語
--- PASS: Test_日本語 (0.00s)
=== RUN   Test_한국어
--- PASS: Test_한국어 (0.00s)
PASS
ok      example.com/test      0.123s
`
	
	_, err = tempFile.WriteString(unicodeContent)
	require.NoError(t, err, "Should write unicode content")
	tempFile.Close()
	
	// 测试分析包含Unicode的文件
	result, err := mst.testAnalyzeTestLog(tempFile.Name())
	assert.NoError(t, err, "Should handle Unicode characters successfully")
	assert.NotNil(t, result, "Result should not be nil")
}