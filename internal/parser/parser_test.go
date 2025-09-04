package parser

import (
	"strings"
	"testing"
	"time"
)

// TestParseTestLog_ValidInput 测试有效输入的解析
func TestParseTestLog_ValidInput(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"=== RUN   TestExample1\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}
{"Time":"2024-01-15T10:30:03Z","Action":"run","Package":"example/pkg","Test":"TestExample2"}
{"Time":"2024-01-15T10:30:04Z","Action":"output","Package":"example/pkg","Test":"TestExample2","Output":"=== RUN   TestExample2\n"}
{"Time":"2024-01-15T10:30:05Z","Action":"output","Package":"example/pkg","Test":"TestExample2","Output":"FAIL: Expected 5, got 3\n"}
{"Time":"2024-01-15T10:30:06Z","Action":"fail","Package":"example/pkg","Test":"TestExample2","Elapsed":0.002}`
	reader := strings.NewReader(testInput)

	// Act
	result, err := ParseTestLog(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalTests != 2 {
		t.Errorf("Expected TotalTests=2, got %d", result.TotalTests)
	}
	if result.PassedTests != 1 {
		t.Errorf("Expected PassedTests=1, got %d", result.PassedTests)
	}
	if result.FailedTests != 1 {
		t.Errorf("Expected FailedTests=1, got %d", result.FailedTests)
	}
	if len(result.PassedTestNames) != 1 || result.PassedTestNames[0] != "TestExample1" {
		t.Errorf("Expected PassedTestNames=[TestExample1], got %v", result.PassedTestNames)
	}
	if len(result.FailedTestNames) != 1 || result.FailedTestNames[0] != "TestExample2" {
		t.Errorf("Expected FailedTestNames=[TestExample2], got %v", result.FailedTestNames)
	}
	if len(result.Packages) != 1 || result.Packages[0] != "example/pkg" {
		t.Errorf("Expected Packages=[example/pkg], got %v", result.Packages)
	}
}

// TestParseTestLog_EmptyInput 测试空输入
func TestParseTestLog_EmptyInput(t *testing.T) {
	// Arrange
	reader := strings.NewReader("")

	// Act
	result, err := ParseTestLog(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalTests != 0 {
		t.Errorf("Expected TotalTests=0, got %d", result.TotalTests)
	}
	if len(result.TestDetails) != 0 {
		t.Errorf("Expected empty TestDetails, got %v", result.TestDetails)
	}
}

// TestParseTestLog_SkippedTests 测试跳过的测试
func TestParseTestLog_SkippedTests(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestSkipped"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestSkipped","Output":"=== RUN   TestSkipped\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"skip","Package":"example/pkg","Test":"TestSkipped","Elapsed":0.001}`
	reader := strings.NewReader(testInput)

	// Act
	result, err := ParseTestLog(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.SkippedTests != 1 {
		t.Errorf("Expected SkippedTests=1, got %d", result.SkippedTests)
	}
	if len(result.SkippedTestNames) != 1 || result.SkippedTestNames[0] != "TestSkipped" {
		t.Errorf("Expected SkippedTestNames=[TestSkipped], got %v", result.SkippedTestNames)
	}
	if detail, exists := result.TestDetails["TestSkipped"]; !exists {
		t.Error("Expected TestSkipped in TestDetails")
	} else if detail.Status != "skip" {
		t.Errorf("Expected status=skip, got %s", detail.Status)
	}
}

// TestParseTestLog_InvalidJSON 测试包含无效JSON的输入
func TestParseTestLog_InvalidJSON(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
invalid json line
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`
	reader := strings.NewReader(testInput)

	// Act
	result, err := ParseTestLog(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.PassedTests != 1 {
		t.Errorf("Expected PassedTests=1, got %d", result.PassedTests)
	}
}

// TestParseTestLog_MultiplePackages 测试多个包的情况
func TestParseTestLog_MultiplePackages(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"pkg1","Test":"Test1"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"pkg1","Test":"Test1","Elapsed":0.001}
{"Time":"2024-01-15T10:30:02Z","Action":"run","Package":"pkg2","Test":"Test2"}
{"Time":"2024-01-15T10:30:03Z","Action":"pass","Package":"pkg2","Test":"Test2","Elapsed":0.002}`
	reader := strings.NewReader(testInput)

	// Act
	result, err := ParseTestLog(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(result.Packages))
	}
	expectedPackages := map[string]bool{"pkg1": true, "pkg2": true}
	for _, pkg := range result.Packages {
		if !expectedPackages[pkg] {
			t.Errorf("Unexpected package: %s", pkg)
		}
	}
}

// TestValidateTestLog_ValidInput 测试有效输入的验证
func TestValidateTestLog_ValidInput(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`
	reader := strings.NewReader(testInput)

	// Act
	err := ValidateTestLog(reader)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestValidateTestLog_EmptyInput 测试空输入的验证
func TestValidateTestLog_EmptyInput(t *testing.T) {
	// Arrange
	reader := strings.NewReader("")

	// Act
	err := ValidateTestLog(reader)

	// Assert
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
	if !strings.Contains(err.Error(), "file is empty") {
		t.Errorf("Expected 'file is empty' error, got %v", err)
	}
}

// TestValidateTestLog_InvalidFormat 测试无效格式的验证
func TestValidateTestLog_InvalidFormat(t *testing.T) {
	// Arrange
	testInput := `not json line 1
not json line 2
not json line 3
not json line 4
not json line 5`
	reader := strings.NewReader(testInput)

	// Act
	err := ValidateTestLog(reader)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
	if !strings.Contains(err.Error(), "does not appear to be go test -json output") {
		t.Errorf("Expected format error, got %v", err)
	}
}

// TestExtractErrorFromOutput 测试错误信息提取
func TestExtractErrorFromOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "FAIL pattern",
			output:   "Some output\nFAIL: Test failed\nMore output",
			expected: "FAIL: Test failed",
		},
		{
			name:     "Error pattern",
			output:   "Error: Something went wrong\nOther line",
			expected: "Error: Something went wrong",
		},
		{
			name:     "Expected/actual pattern",
			output:   "expected: 5\nactual: 3\nother info",
			expected: "expected: 5\nactual: 3",
		},
		{
			name:     "No error patterns",
			output:   "line1\nline2\nline3\nline4\nline5\nline6",
			expected: "line1\nline2\nline3\nline4\nline5",
		},
		{
			name:     "Empty output",
			output:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractErrorFromOutput(tt.output)

			// Assert
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestTestEvent_TimeHandling 测试时间处理
func TestTestEvent_TimeHandling(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00.123456789Z","Action":"run","Package":"example/pkg","Test":"TestTime"}
{"Time":"2024-01-15T10:30:01.987654321Z","Action":"pass","Package":"example/pkg","Test":"TestTime","Elapsed":1.864}`
	reader := strings.NewReader(testInput)

	// Act
	result, err := ParseTestLog(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if detail, exists := result.TestDetails["TestTime"]; !exists {
		t.Error("Expected TestTime in TestDetails")
	} else {
		if detail.Elapsed != 1.864 {
			t.Errorf("Expected elapsed=1.864, got %f", detail.Elapsed)
		}
	}
}

// TestParseTestLog_ConcurrentAccess 测试并发访问安全性
func TestParseTestLog_ConcurrentAccess(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestConcurrent"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"example/pkg","Test":"TestConcurrent","Elapsed":0.001}`

	// Act & Assert
	// 并发执行多次解析，确保没有竞态条件
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			reader := strings.NewReader(testInput)
			result, err := ParseTestLog(reader)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result.PassedTests != 1 {
				t.Errorf("Expected PassedTests=1, got %d", result.PassedTests)
			}
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}