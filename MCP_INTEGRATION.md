# MCP 集成配置指南

本文档说明如何在其他工具中配置和使用 Go Test Reader MCP 服务器。

## 基本配置

### 1. 使用 go run 命令（开发环境）

```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": [
        "run",
        "main.go"
      ],
      "cwd": "/path/to/go_test_reader",
      "env": {
        "GO111MODULE": "on"
      }
    }
  }
}
```

### 2. 使用编译后的可执行文件（生产环境）

首先编译项目：
```bash
go build -o go_test_reader
```

然后配置：
```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "./go_test_reader",
      "args": [],
      "cwd": "/path/to/go_test_reader"
    }
  }
}
```

### 3. Windows 环境配置

```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": [
        "run",
        "main.go"
      ],
      "cwd": "C:\\path\\to\\go_test_reader",
      "env": {
        "GO111MODULE": "on"
      }
    }
  }
}
```

或使用编译后的 exe 文件：
```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go_test_reader.exe",
      "args": [],
      "cwd": "C:\\path\\to\\go_test_reader"
    }
  }
}
```

## 高级配置

### 1. 带环境变量的配置

```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": [
        "run",
        "main.go"
      ],
      "cwd": "/path/to/go_test_reader",
      "env": {
        "GO111MODULE": "on",
        "GOPROXY": "https://proxy.golang.org,direct",
        "GOSUMDB": "sum.golang.org",
        "CGO_ENABLED": "0"
      }
    }
  }
}
```

### 2. 多个 MCP 服务器配置

```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": [
        "run",
        "main.go"
      ],
      "cwd": "/path/to/go_test_reader",
      "env": {
        "GO111MODULE": "on"
      }
    },
    "other-mcp-server": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-server-example"
      ]
    }
  }
}
```

## 可用工具

配置完成后，你可以使用以下 MCP 工具：

### 1. upload_test_log
- **功能**: 上传 go test -json 输出的测试日志文件进行分析
- **参数**: `file_path` - 测试日志文件的绝对路径
- **返回**: 任务ID和状态信息

### 2. get_analysis_result
- **功能**: 根据任务ID获取分析结果
- **参数**: `task_id` - 任务ID
- **返回**: 详细的测试分析结果

### 3. terminate_task
- **功能**: 终止正在运行的分析任务
- **参数**: `task_id` - 要终止的任务ID
- **返回**: 终止操作的状态

### 4. get_test_details
- **功能**: 获取特定测试的详细信息
- **参数**: `task_id` - 任务ID, `test_name` - 测试名称
- **返回**: 测试的详细错误信息和输出

## 使用示例

1. **生成测试日志**:
   ```bash
   go test -json ./... > test_results.json
   ```

2. **上传并分析**:
   使用 `upload_test_log` 工具上传 `test_results.json` 文件

3. **查看结果**:
   使用返回的任务ID调用 `get_analysis_result` 获取分析结果

4. **查看失败测试详情**:
   对于失败的测试，使用 `get_test_details` 获取详细错误信息

## 故障排除

### 常见问题

1. **服务器启动失败**
   - 检查 Go 环境是否正确安装
   - 确认工作目录路径正确
   - 检查依赖是否已下载 (`go mod download`)

2. **文件路径问题**
   - 确保使用绝对路径
   - Windows 用户注意路径分隔符的转义

3. **权限问题**
   - 确保 MCP 客户端有权限访问指定的工作目录
   - 确保有权限读取测试日志文件

### 调试建议

1. 先在命令行中手动运行服务器确认正常工作
2. 检查 MCP 客户端的日志输出
3. 确认测试日志文件格式正确（`go test -json` 输出）

## 注意事项

- 服务器通过标准输入/输出与 MCP 客户端通信
- 支持并发处理多个分析任务
- 任务会自动分配唯一ID进行跟踪
- 服务器会自动清理完成的任务资源