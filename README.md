# Go Test Reader MCP Server

一个基于 Model Context Protocol (MCP) 的 Go 语言跨平台服务器，用于解析和分析 `go test -json` 输出的单元测试日志文件。

## 功能特性

### 核心功能
- 解析 `go test -json` 输出的单元测试日志文件
- 提取关键测试信息：总数量、通过/失败数量、测试名称列表
- 支持通过测试名称查询详细错误信息
- 自动生成唯一任务ID进行任务跟踪

### 服务器特性
- 支持并发处理多个LLM请求
- 提供任务终止接口
- 健壮的文件处理机制
- 跨平台兼容性
- 基于 MCP 协议的标准化接口

## 架构设计

```
go_test_reader/
├── main.go                    # 程序入口
├── internal/
│   ├── server/
│   │   └── server.go         # MCP 服务器实现
│   ├── task/
│   │   └── manager.go        # 任务管理器
│   └── parser/
│       └── parser.go         # 测试日志解析器
├── go.mod
└── README.md
```

### 组件说明

#### 1. MCP Server (`internal/server`)
- 实现 MCP 协议接口
- 注册和处理工具调用
- 管理并发请求
- 协调任务管理器和解析器

#### 2. Task Manager (`internal/task`)
- 任务生命周期管理
- 并发安全的任务状态跟踪
- 支持任务取消和清理
- 提供任务查询接口

#### 3. Parser (`internal/parser`)
- 解析 `go test -json` 格式输出
- 提取测试统计信息
- 收集详细的测试结果和错误信息
- 验证日志文件格式

## MCP 工具接口

### 1. upload_test_log
上传测试日志文件进行分析

**参数：**
- `file_path` (string): 测试日志文件路径

**返回：**
```json
{
  "task_id": "uuid",
  "status": "started",
  "message": "测试日志分析任务已启动"
}
```

### 2. get_analysis_result
根据任务ID获取分析结果

**参数：**
- `task_id` (string): 任务ID

**返回：**
```json
{
  "task_id": "uuid",
  "status": "completed",
  "file_path": "/path/to/test.log",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:01:00Z",
  "result": {
    "total_tests": 10,
    "passed_tests": 8,
    "failed_tests": 2,
    "skipped_tests": 0,
    "failed_test_names": ["TestExample1", "TestExample2"],
    "passed_test_names": ["TestExample3", "TestExample4", ...]
  }
}
```

### 3. terminate_task
终止指定的分析任务

**参数：**
- `task_id` (string): 要终止的任务ID

**返回：**
```json
{
  "task_id": "uuid",
  "status": "terminated",
  "message": "任务已终止"
}
```

### 4. get_test_details
根据任务ID和测试名称获取详细错误信息

**参数：**
- `task_id` (string): 任务ID
- `test_name` (string): 测试名称

**返回：**
```json
{
  "test_name": "TestExample1",
  "status": "fail",
  "output": "=== RUN   TestExample1\n--- FAIL: TestExample1 (0.00s)\n    example_test.go:10: expected 5, got 3\n",
  "error": "expected 5, got 3",
  "elapsed": 0.001
}
```

## 使用方法

### 1. 构建项目
```bash
go build -v
```

### 2. 运行服务器
```bash
./go_test_reader
```

服务器将通过标准输入/输出与 MCP 客户端通信。

### 3. 生成测试日志
```bash
go test -json ./... > test_output.json
```

### 4. 通过 MCP 客户端使用
使用支持 MCP 协议的客户端（如 Claude Desktop）连接到服务器，然后使用提供的工具进行测试日志分析。

## 技术特性

### 并发处理
- 使用 Go 的 goroutine 实现并发任务处理
- 线程安全的任务状态管理
- 支持任务取消和超时处理

### 错误处理
- 健壮的文件读取和解析
- 详细的错误信息收集
- 优雅的错误恢复机制

### 跨平台兼容性
- 纯 Go 实现，支持所有 Go 支持的平台
- 标准库依赖，最小化外部依赖
- 文件路径处理兼容不同操作系统

## 依赖项

- Go 1.23.4+
- github.com/modelcontextprotocol/go-sdk
- github.com/google/uuid

## 许可证

本项目采用 MIT 许可证。