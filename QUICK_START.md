# 快速开始指南

本指南帮助你快速在其他工具中配置和使用 Go Test Reader MCP 服务器。

## 第一步：确认服务器可以运行

在项目目录中运行以下命令确认服务器正常工作：

```bash
cd /path/to/go_test_reader
go run main.go
```

如果看到服务器启动并等待输入，说明配置正确。按 `Ctrl+C` 停止服务器。

## 第二步：配置 MCP 客户端

### 方法一：使用 go run（推荐用于开发）

在你的 MCP 客户端配置文件中添加：

```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": ["run", "main.go"],
      "cwd": "/path/to/go_test_reader",
      "env": {
        "GO111MODULE": "on"
      }
    }
  }
}
```

**Windows 用户请使用：**
```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": ["run", "main.go"],
      "cwd": "C:\\path\\to\\go_test_reader",
      "env": {
        "GO111MODULE": "on"
      }
    }
  }
}
```

### 方法二：使用编译后的可执行文件（推荐用于生产）

1. 编译项目：
   ```bash
   go build -o go_test_reader
   ```

2. 配置 MCP 客户端：
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

## 第三步：生成测试数据

在你的 Go 项目中生成测试日志：

```bash
cd /path/to/your/go/project
go test -json ./... > test_results.json
```

## 第四步：使用 MCP 工具

配置完成后，你可以在 MCP 客户端中使用以下工具：

### 1. 上传测试日志
使用 `upload_test_log` 工具：
- 参数：`file_path` - 测试日志文件的绝对路径
- 返回：任务 ID

### 2. 查看分析结果
使用 `get_analysis_result` 工具：
- 参数：`task_id` - 上一步返回的任务 ID
- 返回：详细的测试分析结果

### 3. 查看失败测试详情
使用 `get_test_details` 工具：
- 参数：`task_id` 和 `test_name`
- 返回：特定测试的详细错误信息

## 示例工作流程

1. **生成测试日志**：
   ```bash
   go test -json ./... > /tmp/test_results.json
   ```

2. **上传并分析**：
   - 调用 `upload_test_log`，参数：`{"file_path": "/tmp/test_results.json"}`
   - 获得任务 ID，例如：`"task-123-456"`

3. **查看结果**：
   - 调用 `get_analysis_result`，参数：`{"task_id": "task-123-456"}`
   - 查看测试统计和失败测试列表

4. **查看详细错误**：
   - 对于失败的测试，调用 `get_test_details`
   - 参数：`{"task_id": "task-123-456", "test_name": "TestFailedExample"}`

## 常见问题

### Q: 服务器启动失败
A: 检查 Go 环境和依赖：
```bash
go version
go mod download
```

### Q: 路径配置问题
A: 确保使用绝对路径，Windows 用户注意转义反斜杠。

### Q: 测试日志格式错误
A: 确保使用 `go test -json` 生成日志，不是普通的 `go test` 输出。

## 更多信息

- 详细配置选项：[MCP_INTEGRATION.md](MCP_INTEGRATION.md)
- 完整使用指南：[USAGE.md](USAGE.md)
- 项目文档：[README.md](README.md)