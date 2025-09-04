package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// TestEvent go test -json 输出的事件结构
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output"`
	Elapsed float64   `json:"Elapsed"`
}

// TestDetail 测试详细信息
type TestDetail struct {
	Status  string  `json:"status"`
	Output  string  `json:"output"`
	Error   string  `json:"error"`
	Elapsed float64 `json:"elapsed"`
}

// TestResult 测试结果汇总
type TestResult struct {
	TotalTests       int                    `json:"total_tests"`
	PassedTests      int                    `json:"passed_tests"`
	FailedTests      int                    `json:"failed_tests"`
	SkippedTests     int                    `json:"skipped_tests"`
	FailedTestNames  []string               `json:"failed_test_names"`
	PassedTestNames  []string               `json:"passed_test_names"`
	SkippedTestNames []string               `json:"skipped_test_names"`
	TestDetails      map[string]*TestDetail `json:"test_details"`
	Packages         []string               `json:"packages"`
}

// ParseTestLog 解析 go test -json 输出
func ParseTestLog(reader io.Reader) (*TestResult, error) {
	result := &TestResult{
		FailedTestNames:  make([]string, 0),
		PassedTestNames:  make([]string, 0),
		SkippedTestNames: make([]string, 0),
		TestDetails:      make(map[string]*TestDetail),
		Packages:         make([]string, 0),
	}
	
	packageSet := make(map[string]bool)
	testOutputs := make(map[string][]string)
	
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// 跳过无法解析的行，可能是非JSON输出
			continue
		}
		
		// 记录包信息
		if event.Package != "" && !packageSet[event.Package] {
			packageSet[event.Package] = true
			result.Packages = append(result.Packages, event.Package)
		}
		
		// 处理测试事件
		if event.Test != "" {
			switch event.Action {
			case "run":
				// 测试开始运行
				if _, exists := result.TestDetails[event.Test]; !exists {
					result.TestDetails[event.Test] = &TestDetail{
						Status: "running",
						Output: "",
					}
				}
				
			case "output":
				// 收集测试输出
				if event.Output != "" {
					testOutputs[event.Test] = append(testOutputs[event.Test], event.Output)
				}
				
			case "pass":
				// 测试通过
				result.PassedTests++
				result.PassedTestNames = append(result.PassedTestNames, event.Test)
				
				if detail, exists := result.TestDetails[event.Test]; exists {
					detail.Status = "pass"
					detail.Elapsed = event.Elapsed
					detail.Output = strings.Join(testOutputs[event.Test], "")
				} else {
					result.TestDetails[event.Test] = &TestDetail{
						Status:  "pass",
						Elapsed: event.Elapsed,
						Output:  strings.Join(testOutputs[event.Test], ""),
					}
				}
				
			case "fail":
				// 测试失败
				result.FailedTests++
				result.FailedTestNames = append(result.FailedTestNames, event.Test)
				
				output := strings.Join(testOutputs[event.Test], "")
				errorMsg := extractErrorFromOutput(output)
				
				if detail, exists := result.TestDetails[event.Test]; exists {
					detail.Status = "fail"
					detail.Elapsed = event.Elapsed
					detail.Output = output
					detail.Error = errorMsg
				} else {
					result.TestDetails[event.Test] = &TestDetail{
						Status:  "fail",
						Elapsed: event.Elapsed,
						Output:  output,
						Error:   errorMsg,
					}
				}
				
			case "skip":
				// 测试跳过
				result.SkippedTests++
				result.SkippedTestNames = append(result.SkippedTestNames, event.Test)
				
				if detail, exists := result.TestDetails[event.Test]; exists {
					detail.Status = "skip"
					detail.Elapsed = event.Elapsed
					detail.Output = strings.Join(testOutputs[event.Test], "")
				} else {
					result.TestDetails[event.Test] = &TestDetail{
						Status:  "skip",
						Elapsed: event.Elapsed,
						Output:  strings.Join(testOutputs[event.Test], ""),
					}
				}
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading test log: %w", err)
	}
	
	// 计算总测试数
	result.TotalTests = result.PassedTests + result.FailedTests + result.SkippedTests
	
	return result, nil
}

// extractErrorFromOutput 从测试输出中提取错误信息
func extractErrorFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	errorLines := make([]string, 0)
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 查找常见的错误模式
		if strings.Contains(trimmed, "FAIL:") ||
			strings.Contains(trimmed, "Error:") ||
			strings.Contains(trimmed, "panic:") ||
			strings.Contains(trimmed, "expected") ||
			strings.Contains(trimmed, "actual") ||
			strings.Contains(trimmed, "got") ||
			strings.Contains(trimmed, "want") {
			errorLines = append(errorLines, trimmed)
		}
	}
	
	if len(errorLines) > 0 {
		return strings.Join(errorLines, "\n")
	}
	
	// 如果没有找到特定的错误模式，返回整个输出的前几行
	if len(lines) > 0 {
		maxLines := 5
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		return strings.Join(lines[:maxLines], "\n")
	}
	
	return "No error details available"
}

// ValidateTestLog 验证测试日志格式
func ValidateTestLog(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	lineCount := 0
	validLines := 0
	
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			validLines++
		}
		
		// 只检查前100行来判断格式
		if lineCount >= 100 {
			break
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	
	if lineCount == 0 {
		return fmt.Errorf("file is empty")
	}
	
	// 如果有效JSON行少于总行数的50%，可能不是正确的格式
	if validLines < lineCount/2 {
		return fmt.Errorf("file does not appear to be go test -json output (valid JSON lines: %d/%d)", validLines, lineCount)
	}
	
	return nil
}