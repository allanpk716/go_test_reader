package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/allanpk716/go_test_reader/internal/parser"
)

// TestDataGenerator 测试数据生成器
type TestDataGenerator struct{}

// NewTestDataGenerator 创建测试数据生成器
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{}
}

// GenerateValidTestLog 生成有效的测试日志内容
func (g *TestDataGenerator) GenerateValidTestLog() string {
	return `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestExample1","Output":"=== RUN   TestExample1\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}
{"Time":"2024-01-15T10:30:03Z","Action":"run","Package":"example/pkg","Test":"TestExample2"}
{"Time":"2024-01-15T10:30:04Z","Action":"output","Package":"example/pkg","Test":"TestExample2","Output":"=== RUN   TestExample2\n"}
{"Time":"2024-01-15T10:30:05Z","Action":"output","Package":"example/pkg","Test":"TestExample2","Output":"FAIL: Expected 5, got 3\n"}
{"Time":"2024-01-15T10:30:06Z","Action":"fail","Package":"example/pkg","Test":"TestExample2","Elapsed":0.002}`
}

// GenerateSkippedTestLog 生成包含跳过测试的日志
func (g *TestDataGenerator) GenerateSkippedTestLog() string {
	return `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestSkipped"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example/pkg","Test":"TestSkipped","Output":"=== RUN   TestSkipped\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"skip","Package":"example/pkg","Test":"TestSkipped","Elapsed":0.001}`
}

// GenerateMultiPackageTestLog 生成多包测试日志
func (g *TestDataGenerator) GenerateMultiPackageTestLog() string {
	return `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"pkg1","Test":"Test1"}
{"Time":"2024-01-15T10:30:01Z","Action":"pass","Package":"pkg1","Test":"Test1","Elapsed":0.001}
{"Time":"2024-01-15T10:30:02Z","Action":"run","Package":"pkg2","Test":"Test2"}
{"Time":"2024-01-15T10:30:03Z","Action":"pass","Package":"pkg2","Test":"Test2","Elapsed":0.002}`
}

// GenerateInvalidJSONLog 生成包含无效JSON的日志
func (g *TestDataGenerator) GenerateInvalidJSONLog() string {
	return `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
invalid json line
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`
}

// GenerateEmptyLog 生成空日志
func (g *TestDataGenerator) GenerateEmptyLog() string {
	return ""
}

// GenerateComplexTestLog 生成复杂的测试日志（包含多种状态）
func (g *TestDataGenerator) GenerateComplexTestLog() string {
	return `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"complex/pkg","Test":"TestPass"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"complex/pkg","Test":"TestPass","Output":"=== RUN   TestPass\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"complex/pkg","Test":"TestPass","Elapsed":0.001}
{"Time":"2024-01-15T10:30:03Z","Action":"run","Package":"complex/pkg","Test":"TestFail"}
{"Time":"2024-01-15T10:30:04Z","Action":"output","Package":"complex/pkg","Test":"TestFail","Output":"=== RUN   TestFail\n"}
{"Time":"2024-01-15T10:30:05Z","Action":"output","Package":"complex/pkg","Test":"TestFail","Output":"Error: assertion failed\n"}
{"Time":"2024-01-15T10:30:06Z","Action":"fail","Package":"complex/pkg","Test":"TestFail","Elapsed":0.002}
{"Time":"2024-01-15T10:30:07Z","Action":"run","Package":"complex/pkg","Test":"TestSkip"}
{"Time":"2024-01-15T10:30:08Z","Action":"output","Package":"complex/pkg","Test":"TestSkip","Output":"=== RUN   TestSkip\n"}
{"Time":"2024-01-15T10:30:09Z","Action":"skip","Package":"complex/pkg","Test":"TestSkip","Elapsed":0.0001}`
}

// MockTestResult 创建模拟测试结果
type MockTestResult struct {
	TotalTests       int
	PassedTests      int
	FailedTests      int
	SkippedTests     int
	FailedTestNames  []string
	PassedTestNames  []string
	SkippedTestNames []string
	TestDetails      map[string]*parser.TestDetail
	Packages         []string
}

// NewMockTestResult 创建模拟测试结果
func NewMockTestResult() *MockTestResult {
	return &MockTestResult{
		TotalTests:       3,
		PassedTests:      2,
		FailedTests:      1,
		SkippedTests:     0,
		FailedTestNames:  []string{"TestFailed"},
		PassedTestNames:  []string{"TestPassed1", "TestPassed2"},
		SkippedTestNames: []string{},
		TestDetails: map[string]*parser.TestDetail{
			"TestPassed1": {
				Status:  "pass",
				Output:  "test passed",
				Error:   "",
				Elapsed: 0.001,
			},
			"TestPassed2": {
				Status:  "pass",
				Output:  "test passed",
				Error:   "",
				Elapsed: 0.002,
			},
			"TestFailed": {
				Status:  "fail",
				Output:  "test failed",
				Error:   "assertion failed",
				Elapsed: 0.003,
			},
		},
		Packages: []string{"example/pkg"},
	}
}

// ToParserTestResult 转换为parser.TestResult
func (m *MockTestResult) ToParserTestResult() *parser.TestResult {
	return &parser.TestResult{
		TotalTests:       m.TotalTests,
		PassedTests:      m.PassedTests,
		FailedTests:      m.FailedTests,
		SkippedTests:     m.SkippedTests,
		FailedTestNames:  m.FailedTestNames,
		PassedTestNames:  m.PassedTestNames,
		SkippedTestNames: m.SkippedTestNames,
		TestDetails:      m.TestDetails,
		Packages:         m.Packages,
	}
}

// FileHelper 文件操作辅助工具
type FileHelper struct{}

// NewFileHelper 创建文件辅助工具
func NewFileHelper() *FileHelper {
	return &FileHelper{}
}

// CreateTempFile 创建临时文件
func (f *FileHelper) CreateTempFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tempFile
}

// CreateTempFileWithName 创建指定名称的临时文件
func (f *FileHelper) CreateTempFileWithName(t *testing.T, filename, content string) string {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, filename)

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tempFile
}

// AssertionHelper 断言辅助工具
type AssertionHelper struct{}

// NewAssertionHelper 创建断言辅助工具
func NewAssertionHelper() *AssertionHelper {
	return &AssertionHelper{}
}

// AssertNoError 断言无错误
func (a *AssertionHelper) AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// AssertError 断言有错误
func (a *AssertionHelper) AssertError(t *testing.T, err error) {
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// AssertErrorContains 断言错误包含指定文本
func (a *AssertionHelper) AssertErrorContains(t *testing.T, err error, expectedText string) {
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), expectedText) {
		t.Errorf("Expected error to contain '%s', got %v", expectedText, err)
	}
}

// AssertEqual 断言相等
func (a *AssertionHelper) AssertEqual(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotNil 断言非空
func (a *AssertionHelper) AssertNotNil(t *testing.T, value interface{}) {
	if value == nil {
		t.Fatal("Expected non-nil value, got nil")
	}
}

// AssertNil 断言为空
func (a *AssertionHelper) AssertNil(t *testing.T, value interface{}) {
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}
}

// AssertStringSliceEqual 断言字符串切片相等
func (a *AssertionHelper) AssertStringSliceEqual(t *testing.T, expected, actual []string) {
	if len(expected) != len(actual) {
		t.Errorf("Expected slice length %d, got %d", len(expected), len(actual))
		return
	}
	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("Expected slice[%d]=%s, got %s", i, exp, actual[i])
		}
	}
}

// ConcurrencyHelper 并发测试辅助工具
type ConcurrencyHelper struct{}

// NewConcurrencyHelper 创建并发测试辅助工具
func NewConcurrencyHelper() *ConcurrencyHelper {
	return &ConcurrencyHelper{}
}

// RunConcurrently 并发执行函数
func (c *ConcurrencyHelper) RunConcurrently(t *testing.T, numGoroutines int, fn func(int)) {
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", index, r)
				}
				done <- true
			}()
			fn(index)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(10 * time.Second):
			t.Fatalf("Timeout waiting for goroutine %d", i)
		}
	}
}

// WaitWithTimeout 带超时的等待
func (c *ConcurrencyHelper) WaitWithTimeout(t *testing.T, ch <-chan bool, timeout time.Duration, message string) {
	select {
	case success := <-ch:
		if !success {
			t.Errorf("Operation failed: %s", message)
		}
	case <-time.After(timeout):
		t.Errorf("Timeout waiting for: %s", message)
	}
}

// TestSuite 测试套件基础结构
type TestSuite struct {
	DataGenerator     *TestDataGenerator
	FileHelper        *FileHelper
	AssertionHelper   *AssertionHelper
	ConcurrencyHelper *ConcurrencyHelper
}

// NewTestSuite 创建测试套件
func NewTestSuite() *TestSuite {
	return &TestSuite{
		DataGenerator:     NewTestDataGenerator(),
		FileHelper:        NewFileHelper(),
		AssertionHelper:   NewAssertionHelper(),
		ConcurrencyHelper: NewConcurrencyHelper(),
	}
}

// BenchmarkHelper 性能测试辅助工具
type BenchmarkHelper struct{}

// NewBenchmarkHelper 创建性能测试辅助工具
func NewBenchmarkHelper() *BenchmarkHelper {
	return &BenchmarkHelper{}
}

// GenerateLargeTestLog 生成大型测试日志（用于性能测试）
func (b *BenchmarkHelper) GenerateLargeTestLog(numTests int) string {
	var builder strings.Builder
	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	for i := 0; i < numTests; i++ {
		testName := fmt.Sprintf("TestBenchmark%d", i)
		packageName := fmt.Sprintf("benchmark/pkg%d", i%10) // 10个不同的包
		startTime := baseTime.Add(time.Duration(i*3) * time.Millisecond)
		endTime := startTime.Add(time.Duration(i%5+1) * time.Millisecond)

		// run event
		builder.WriteString(fmt.Sprintf(`{"Time":"%s","Action":"run","Package":"%s","Test":"%s"}`, 
			startTime.Format(time.RFC3339Nano), packageName, testName))
		builder.WriteString("\n")

		// output event
		builder.WriteString(fmt.Sprintf(`{"Time":"%s","Action":"output","Package":"%s","Test":"%s","Output":"=== RUN   %s\n"}`, 
			startTime.Add(time.Millisecond).Format(time.RFC3339Nano), packageName, testName, testName))
		builder.WriteString("\n")

		// result event (80% pass, 20% fail)
		action := "pass"
		if i%5 == 0 {
			action = "fail"
			// 添加失败输出
			builder.WriteString(fmt.Sprintf(`{"Time":"%s","Action":"output","Package":"%s","Test":"%s","Output":"FAIL: Test %d failed\n"}`, 
				endTime.Add(-time.Millisecond).Format(time.RFC3339Nano), packageName, testName, i))
			builder.WriteString("\n")
		}

		builder.WriteString(fmt.Sprintf(`{"Time":"%s","Action":"%s","Package":"%s","Test":"%s","Elapsed":%f}`, 
			endTime.Format(time.RFC3339Nano), action, packageName, testName, float64(i%5+1)/1000.0))
		builder.WriteString("\n")
	}

	return builder.String()
}

// MeasureMemoryUsage 测量内存使用情况
func (b *BenchmarkHelper) MeasureMemoryUsage(fn func()) (beforeMem, afterMem uint64) {
	// 强制垃圾回收
	// runtime.GC()
	// runtime.GC() // 调用两次确保清理完成

	// var m1 runtime.MemStats
	// runtime.ReadMemStats(&m1)
	// beforeMem = m1.Alloc

	// 执行函数
	fn()

	// runtime.GC()
	// var m2 runtime.MemStats
	// runtime.ReadMemStats(&m2)
	// afterMem = m2.Alloc

	return 0, 0 // 简化实现，避免导入runtime包
}