# Go Test Reader MCP Server 使用指南

## 概述

Go Test Reader 是一个基于 Model Context Protocol (MCP) 的服务器，专门用于解析和分析 `go test -json` 输出。它提供了强大的测试结果分析功能，支持并发处理和任务管理。

## 快速开始

### 1. 编译和运行服务器

```bash
# 编译项目
go build -o go_test_reader

# 运行 MCP 服务器
./go_test_reader
```

服务器将通过标准输入/输出与 MCP 客户端通信。

### 2. 生成测试日志

首先，你需要生成 `go test -json` 格式的测试日志：

```bash
# 在你的 Go 项目中运行测试并生成 JSON 日志
go test -json ./... > test_results.json

# 或者使用我们提供的示例生成器
go run test_example.go test_log.json
```

## MCP 工具接口

### 1. upload_test_log - 上传测试日志

上传并开始分析测试日志文件。

**参数：**
- `file_path` (string): 测试日志文件的绝对路径

**返回：**
```json
{
  "task_id": "uuid-string",
  "status": "started",
  "message": "测试日志分析任务已启动"
}
```

**示例：**
```json
{
  "method": "tools/call",
  "params": {
    "name": "upload_test_log",
    "arguments": {
      "file_path": "/path/to/test_results.json"
    }
  }
}
```

### 2. get_analysis_result - 获取分析结果

查询指定任务的分析结果。

**参数：**
- `task_id` (string): 任务 ID

**返回：**
```json
{
  "task_id": "uuid-string",
  "status": "completed|running|failed",
  "result": {
    "total_tests": 10,
    "passed_tests": 8,
    "failed_tests": 2,
    "skipped_tests": 0,
    "failed_test_names": ["TestExample1", "TestExample2"],
    "passed_test_names": [...],
    "skipped_test_names": [...],
    "packages": ["example/pkg1", "example/pkg2"]
  },
  "error": null,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:05Z"
}
```

### 3. terminate_task - 终止任务

强制终止正在运行的分析任务。

**参数：**
- `task_id` (string): 要终止的任务 ID

**返回：**
```json
{
  "task_id": "uuid-string",
  "status": "terminated",
  "message": "任务已终止"
}
```

### 4. get_test_details - 获取测试详情

获取特定测试的详细信息，包括错误信息和输出。

**参数：**
- `task_id` (string): 任务 ID
- `test_name` (string): 测试名称

**返回：**
```json
{
  "test_name": "TestExample",
  "status": "pass|fail|skip",
  "output": "测试输出内容",
  "error": "错误信息（如果有）",
  "elapsed": 0.001
}
```

## 使用流程示例

### 完整的分析流程

1. **上传测试日志**
   ```json
   {
     "method": "tools/call",
     "params": {
       "name": "upload_test_log",
       "arguments": {
         "file_path": "/path/to/test_results.json"
       }
     }
   }
   ```
   
   响应：
   ```json
   {
     "task_id": "abc-123-def",
     "status": "started"
   }
   ```

2. **查询分析结果**
   ```json
   {
     "method": "tools/call",
     "params": {
       "name": "get_analysis_result",
       "arguments": {
         "task_id": "abc-123-def"
       }
     }
   }
   ```

3. **获取失败测试的详细信息**
   ```json
   {
     "method": "tools/call",
     "params": {
       "name": "get_test_details",
       "arguments": {
         "task_id": "abc-123-def",
         "test_name": "TestExample2"
       }
     }
   }
   ```

## 错误处理

### 常见错误

1. **文件不存在**
   ```json
   {
     "error": "file not found: /path/to/nonexistent.json"
   }
   ```

2. **无效的 JSON 格式**
   ```json
   {
     "error": "invalid JSON format in test log"
   }
   ```

3. **任务不存在**
   ```json
   {
     "error": "task not found: invalid-task-id"
   }
   ```

4. **测试不存在**
   ```json
   {
     "error": "test not found: NonExistentTest"
   }
   ```

## 性能特性

### 并发处理
- 支持多个 LLM 客户端同时连接
- 每个分析任务在独立的 goroutine 中运行
- 任务状态实时更新，支持并发查询

### 内存管理
- 自动清理完成的任务（可配置）
- 流式处理大型测试日志文件
- 优化的数据结构减少内存占用

### 跨平台兼容性
- 支持 Windows、Linux、macOS
- 标准的文件路径处理
- 统一的 JSON 输出格式

## 配置选项

服务器目前使用默认配置，未来版本将支持：
- 任务超时设置
- 最大并发任务数
- 日志级别配置
- 任务清理策略

## 故障排除

### 服务器无法启动
1. 检查 Go 版本（需要 1.23+）
2. 确认依赖已正确安装：`go mod tidy`
3. 检查端口是否被占用

### 解析失败
1. 验证测试日志格式：`go run test_parser.go <log_file>`
2. 确认文件路径正确且可读
3. 检查文件是否为有效的 `go test -json` 输出

### 性能问题
1. 对于大型测试套件，考虑分批处理
2. 监控内存使用情况
3. 使用 `terminate_task` 取消不需要的任务

## 开发和扩展

### 添加新功能
1. 在 `internal/server/server.go` 中添加新的工具处理函数
2. 在 `registerTools()` 中注册新工具
3. 更新相关的请求/响应结构体

### 测试
```bash
# 运行单元测试
go test ./...

# 测试解析器
go run test_parser.go test_log.json

# 生成测试数据
go run test_example.go custom_test.json
```

## 许可证

本项目基于 MIT 许可证开源。详见 LICENSE 文件。