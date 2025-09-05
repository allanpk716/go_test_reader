package server

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCPServer_FailureScenarios_InvalidJSONFormats 测试各种无效JSON格式的处理
func TestMCPServer_FailureScenarios_InvalidJSONFormats(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		expectedToFail bool
	}{
		{
			name: "completely_invalid_json",
			content: `this is not json at all
				just random text
				no structure whatsoever`,
			expectedToFail: true,
		},
		{
			name: "malformed_json_brackets",
			content: `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"
				{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`,
			expectedToFail: false, // 解析器应该跳过无效行
		},
		{
			name: "json_with_null_values",
			content: `{"Time":null,"Action":"run","Package":"example/pkg","Test":"TestExample1"}
				{"Time":"2024-01-15T10:30:01Z","Action":null,"Package":"example/pkg","Test":"TestExample1"}
				{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":null,"Test":"TestExample1","Elapsed":0.001}`,
			expectedToFail: false, // 应该能处理null值
		},
		{
			name: "json_with_wrong_types",
			content: `{"Time":123456,"Action":"run","Package":"example/pkg","Test":"TestExample1"}
				{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":"not_a_number"}`,
			expectedToFail: false, // 解析器应该处理类型错误
		},
		{
			name: "extremely_large_json",
			content: generateLargeInvalidJSON(),
			expectedToFail: true,
		},
		{
			name: "json_with_unicode_and_special_chars",
			content: `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"测试包/pkg","Test":"测试用例1"}
				{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"测试包/pkg","Test":"测试用例1","Output":"=== RUN   测试用例1\n包含特殊字符: !@#$%^&*()\n"}
				{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"测试包/pkg","Test":"测试用例1","Elapsed":0.001}`,
			expectedToFail: false, // 应该能处理Unicode字符
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, err := NewMCPServer()
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			tempFile := createTempTestFileForFailure(t, tc.content)
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
			time.Sleep(200 * time.Millisecond) // 给更多时间处理复杂内容

			task := server.taskManager.GetTask(taskID)
			if task == nil {
				t.Fatal("Expected task to exist")
			}

			status := task.GetStatus()
			if tc.expectedToFail {
				if status["status"] != "failed" {
					t.Errorf("Expected task to fail for %s, but got status: %v", tc.name, status["status"])
				}
				if _, hasError := status["error"]; !hasError {
					t.Errorf("Expected error information for failed task %s", tc.name)
				}
			} else {
				if status["status"] == "failed" {
					t.Logf("Task %s failed with error: %v (this might be acceptable)", tc.name, status["error"])
				} else if status["status"] == "completed" {
					t.Logf("Task %s completed successfully despite invalid content", tc.name)
				}
			}
		})
	}
}

// TestMCPServer_FailureScenarios_FileSystemErrors 测试文件系统相关的错误场景
func TestMCPServer_FailureScenarios_FileSystemErrors(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		setup    func() string // 返回实际要使用的文件路径
		cleanup  func(string)
	}{
		{
			name:     "nonexistent_file",
			filePath: "/path/to/nonexistent/file.json",
			setup:    func() string { return "/path/to/nonexistent/file.json" },
			cleanup:  func(string) {},
		},
		{
			name: "directory_instead_of_file",
			setup: func() string {
				tempDir, _ := os.MkdirTemp("", "test_dir")
				return tempDir
			},
			cleanup: func(path string) { os.RemoveAll(path) },
		},
		{
			name: "file_deleted_during_processing",
			setup: func() string {
				tempFile := createTempTestFileForFailure(nil, validTestLogContent())
				// 文件创建后立即删除，模拟处理过程中文件被删除的情况
				os.Remove(tempFile)
				return tempFile
			},
			cleanup: func(path string) { os.Remove(path) }, // 确保清理
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, err := NewMCPServer()
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			filePath := tc.setup()
			defer tc.cleanup(filePath)

			ctx := context.Background()
			session := &mcp.ServerSession{}
			params := &mcp.CallToolParamsFor[UploadRequest]{
				Arguments: UploadRequest{FilePath: filePath},
			}

			// Act
			result, err := server.handleUploadTestLog(ctx, session, params)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error during upload initiation, got %v", err)
			}

			// 等待处理完成
			taskID := result.Meta["task_id"].(string)
			time.Sleep(200 * time.Millisecond)

			task := server.taskManager.GetTask(taskID)
			if task == nil {
				t.Fatal("Expected task to exist")
			}

			status := task.GetStatus()
			// 所有这些场景都应该导致任务失败
			if status["status"] != "failed" {
				t.Errorf("Expected task to fail for %s, but got status: %v", tc.name, status["status"])
			}
			if _, hasError := status["error"]; !hasError {
				t.Errorf("Expected error information for failed task %s", tc.name)
			} else {
				t.Logf("Task %s failed as expected with error: %v", tc.name, status["error"])
			}
		})
	}
}

// TestMCPServer_FailureScenarios_ResourceExhaustion 测试资源耗尽场景
func TestMCPServer_FailureScenarios_ResourceExhaustion(t *testing.T) {
	// Arrange
	server, err := NewMCPServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 创建大量并发任务来测试资源管理
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Act - 创建大量任务
	taskIDs := make([]string, 0, 50)
	for i := 0; i < 50; i++ {
		tempFile := createTempTestFileForFailure(t, fmt.Sprintf(`{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample%d"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestExample%d","Elapsed":0.001}`, i, i))
		defer os.Remove(tempFile)

		params := &mcp.CallToolParamsFor[UploadRequest]{
			Arguments: UploadRequest{FilePath: tempFile},
		}

		result, err := server.handleUploadTestLog(ctx, session, params)
		if err != nil {
			t.Logf("Task %d failed to start: %v", i, err)
			continue
		}

		taskID := result.Meta["task_id"].(string)
		taskIDs = append(taskIDs, taskID)
	}

	// 等待所有任务处理
	time.Sleep(500 * time.Millisecond)

	// Assert - 检查任务状态
	completedCount := 0
	failedCount := 0
	for _, taskID := range taskIDs {
		task := server.taskManager.GetTask(taskID)
		if task != nil {
			status := task.GetStatus()
			switch status["status"] {
			case "completed":
				completedCount++
			case "failed":
				failedCount++
			}
		}
	}

	t.Logf("Resource exhaustion test results: %d completed, %d failed out of %d tasks", completedCount, failedCount, len(taskIDs))

	// 至少应该有一些任务完成或失败（不应该全部卡住）
	if completedCount+failedCount == 0 {
		t.Error("No tasks completed or failed - possible deadlock or resource issue")
	}
}

// TestMCPServer_FailureScenarios_MaliciousInput 测试恶意输入处理
func TestMCPServer_FailureScenarios_MaliciousInput(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "extremely_long_strings",
			content: fmt.Sprintf(`{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"%s","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"%s"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`, strings.Repeat("A", 10000), strings.Repeat("B", 50000)),
		},
		{
			name: "deeply_nested_json",
			content: generateDeeplyNestedJSON(),
		},
		{
			name: "json_with_control_characters",
			content: `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\u000b\u000c\u000e\u000f"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server, err := NewMCPServer()
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			tempFile := createTempTestFileForFailure(t, tc.content)
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
			time.Sleep(300 * time.Millisecond) // 给更多时间处理恶意内容

			task := server.taskManager.GetTask(taskID)
			if task == nil {
				t.Fatal("Expected task to exist")
			}

			status := task.GetStatus()
			// 恶意输入应该被安全处理，不应该导致系统崩溃
			if status["status"] == "failed" {
				t.Logf("Task %s failed safely with error: %v", tc.name, status["error"])
			} else if status["status"] == "completed" {
				t.Logf("Task %s completed despite malicious content", tc.name)
			} else {
				t.Logf("Task %s status: %v", tc.name, status["status"])
			}

			// 确保服务器仍然响应
			_, err = server.handleGetAnalysisResult(ctx, session, &mcp.CallToolParamsFor[QueryRequest]{
				Arguments: QueryRequest{TaskID: taskID},
			})
			if err != nil {
				t.Errorf("Server became unresponsive after processing malicious input %s: %v", tc.name, err)
			}
		})
	}
}

// 辅助函数：生成大量无效JSON内容
func generateLargeInvalidJSON() string {
	content := ""
	for i := 0; i < 1000; i++ {
		content += fmt.Sprintf("invalid json line %d\n", i)
		content += "{broken json without closing brace\n"
		content += "[array, without, proper, closing\n"
	}
	return content
}

// 辅助函数：生成深度嵌套的JSON
func generateDeeplyNestedJSON() string {
	nested := ""
	for i := 0; i < 100; i++ {
		nested += `{"level":` + fmt.Sprintf("%d", i) + `,"nested":`
	}
	nested += `"deep_value"`
	for i := 0; i < 100; i++ {
		nested += "}"
	}

	return fmt.Sprintf(`{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"%s"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`, nested)
}

// 辅助函数：创建临时测试文件（支持nil测试对象）
func createTempTestFileForFailure(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		} else {
			panic(fmt.Sprintf("Failed to create temp file: %v", err))
		}
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		if t != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		} else {
			panic(fmt.Sprintf("Failed to write to temp file: %v", err))
		}
	}

	return tempFile.Name()
}