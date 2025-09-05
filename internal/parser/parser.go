package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
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

// ParseTestTextLog 解析 go test 普通文本输出
func ParseTestTextLog(reader io.Reader) (*TestResult, error) {
	result := &TestResult{
		FailedTestNames:  make([]string, 0),
		PassedTestNames:  make([]string, 0),
		SkippedTestNames: make([]string, 0),
		TestDetails:      make(map[string]*TestDetail),
		Packages:         make([]string, 0),
	}
	
	packageSet := make(map[string]bool)
	currentTest := ""
	currentOutput := make([]string, 0)
	buildErrors := make([]string, 0)
	
	// 正则表达式模式
	runPattern := regexp.MustCompile(`^=== RUN\s+(.+)$`)
	passPattern := regexp.MustCompile(`^--- PASS:\s+(.+?)\s+\(([0-9.]+)s\)$`)
	failPattern := regexp.MustCompile(`^--- FAIL:\s+(.+?)\s+\(([0-9.]+)s\)$`)
	skipPattern := regexp.MustCompile(`^--- SKIP:\s+(.+?)\s+\(([0-9.]+)s\)$`)
	okPattern := regexp.MustCompile(`^(ok|PASS)\s+(.+?)(?:\s+\(cached\))?(?:\s+([0-9.]+)s)?$`)
	failPackagePattern := regexp.MustCompile(`^FAIL\s+(.+?)(?:\s+\[build failed\])?(?:\s+([0-9.]+)s)?$`)
	buildErrorPattern := regexp.MustCompile(`^(.+?):\d+:\d+:\s+(.+)$`)
	
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "" {
			continue
		}
		
		// 检查是否是测试运行开始
		if matches := runPattern.FindStringSubmatch(trimmed); matches != nil {
			// 保存之前测试的输出
			if currentTest != "" && len(currentOutput) > 0 {
				if detail, exists := result.TestDetails[currentTest]; exists {
					detail.Output = strings.Join(currentOutput, "\n")
				}
			}
			
			currentTest = matches[1]
			currentOutput = make([]string, 0)
			
			// 创建测试详情
			result.TestDetails[currentTest] = &TestDetail{
				Status: "running",
				Output: "",
			}
			continue
		}
		
		// 检查测试通过
		if matches := passPattern.FindStringSubmatch(trimmed); matches != nil {
			testName := matches[1]
			elapsed, _ := strconv.ParseFloat(matches[2], 64)
			
			result.PassedTests++
			result.PassedTestNames = append(result.PassedTestNames, testName)
			
			if detail, exists := result.TestDetails[testName]; exists {
				detail.Status = "pass"
				detail.Elapsed = elapsed
				detail.Output = strings.Join(currentOutput, "\n")
			} else {
				result.TestDetails[testName] = &TestDetail{
					Status:  "pass",
					Elapsed: elapsed,
					Output:  strings.Join(currentOutput, "\n"),
				}
			}
			currentTest = ""
			currentOutput = make([]string, 0)
			continue
		}
		
		// 检查测试失败
		if matches := failPattern.FindStringSubmatch(trimmed); matches != nil {
			testName := matches[1]
			elapsed, _ := strconv.ParseFloat(matches[2], 64)
			
			result.FailedTests++
			result.FailedTestNames = append(result.FailedTestNames, testName)
			
			output := strings.Join(currentOutput, "\n")
			errorMsg := extractErrorFromOutput(output)
			
			if detail, exists := result.TestDetails[testName]; exists {
				detail.Status = "fail"
				detail.Elapsed = elapsed
				detail.Output = output
				detail.Error = errorMsg
			} else {
				result.TestDetails[testName] = &TestDetail{
					Status:  "fail",
					Elapsed: elapsed,
					Output:  output,
					Error:   errorMsg,
				}
			}
			currentTest = ""
			currentOutput = make([]string, 0)
			continue
		}
		
		// 检查测试跳过
		if matches := skipPattern.FindStringSubmatch(trimmed); matches != nil {
			testName := matches[1]
			elapsed, _ := strconv.ParseFloat(matches[2], 64)
			
			result.SkippedTests++
			result.SkippedTestNames = append(result.SkippedTestNames, testName)
			
			if detail, exists := result.TestDetails[testName]; exists {
				detail.Status = "skip"
				detail.Elapsed = elapsed
				detail.Output = strings.Join(currentOutput, "\n")
			} else {
				result.TestDetails[testName] = &TestDetail{
					Status:  "skip",
					Elapsed: elapsed,
					Output:  strings.Join(currentOutput, "\n"),
				}
			}
			currentTest = ""
			currentOutput = make([]string, 0)
			continue
		}
		
		// 检查包测试成功
		if matches := okPattern.FindStringSubmatch(trimmed); matches != nil {
			packageName := matches[2]
			if !packageSet[packageName] {
				packageSet[packageName] = true
				result.Packages = append(result.Packages, packageName)
			}
			continue
		}
		
		// 检查包测试失败
		if matches := failPackagePattern.FindStringSubmatch(trimmed); matches != nil {
			packageName := matches[1]
			if !packageSet[packageName] {
				packageSet[packageName] = true
				result.Packages = append(result.Packages, packageName)
			}
			continue
		}
		
		// 检查编译错误
		if matches := buildErrorPattern.FindStringSubmatch(trimmed); matches != nil {
			buildErrors = append(buildErrors, trimmed)
			continue
		}

		// 检查是否只是 "FAIL" 行
		if trimmed == "FAIL" {
			continue
		}
		
		// 收集当前测试的输出
		if currentTest != "" {
			currentOutput = append(currentOutput, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading test log: %w", err)
	}
	
	// 处理最后一个测试的输出
	if currentTest != "" && len(currentOutput) > 0 {
		if detail, exists := result.TestDetails[currentTest]; exists {
			detail.Output = strings.Join(currentOutput, "\n")
		}
	}
	
	// 如果有编译错误，创建一个特殊的失败测试
	if len(buildErrors) > 0 {
		buildErrorTest := "BuildError"
		result.FailedTests++
		result.FailedTestNames = append(result.FailedTestNames, buildErrorTest)
		result.TestDetails[buildErrorTest] = &TestDetail{
			Status: "fail",
			Output: strings.Join(buildErrors, "\n"),
			Error:  "Build failed",
		}
	}
	
	// 计算总测试数
	result.TotalTests = result.PassedTests + result.FailedTests + result.SkippedTests
	
	return result, nil
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

// ValidateTestTextLog 验证文本格式测试日志
func ValidateTestTextLog(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	lineCount := 0
	testLines := 0
	
	// 文本格式的特征模式
	runPattern := regexp.MustCompile(`^=== RUN\s+`)
	passPattern := regexp.MustCompile(`^--- PASS:\s+`)
	failPattern := regexp.MustCompile(`^--- FAIL:\s+`)
	okPattern := regexp.MustCompile(`^(ok|PASS)\s+`)
	failPackagePattern := regexp.MustCompile(`^FAIL\s+`)
	
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		// 检查是否包含测试相关的模式
		if runPattern.MatchString(line) ||
			passPattern.MatchString(line) ||
			failPattern.MatchString(line) ||
			okPattern.MatchString(line) ||
			failPackagePattern.MatchString(line) ||
			line == "FAIL" {
			testLines++
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
	
	// 如果测试相关行少于总行数的10%，可能不是正确的格式
	if testLines == 0 {
		return fmt.Errorf("file does not appear to be go test text output (no test patterns found)")
	}
	
	return nil
}