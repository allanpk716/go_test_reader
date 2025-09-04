# MCP 配置示例

本文档提供了在其他工具中配置 Go Test Reader MCP 服务器的具体示例。

## 示例 1：使用 go run 命令（开发环境推荐）

```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go",
      "args": [
        "run",
        "main.go"
      ],
      "cwd": "c:\\WorkSpace\\Go2Hell\\src\\github.com\\allanpk716\\go_test_reader",
      "env": {
        "GO111MODULE": "on",
        "GOPROXY": "https://proxy.golang.org,direct",
        "GOSUMDB": "sum.golang.org"
      }
    }
  }
}
```

**特点：**
- 适用于开发环境
- 无需预编译
- 包含 Go 代理配置
- 使用当前项目的绝对路径

## 示例 2：使用编译后的可执行文件（生产环境推荐）

**第一步：编译项目**
```bash
go build -o go_test_reader.exe
```

**第二步：配置 MCP 客户端**
```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go_test_reader.exe",
      "args": [],
      "cwd": "c:\\WorkSpace\\Go2Hell\\src\\github.com\\allanpk716\\go_test_reader"
    }
  }
}
```

**特点：**
- 适用于生产环境
- 启动速度更快
- 不依赖 Go 运行时环境
- 配置更简洁

## 可用的 MCP 工具

配置完成后，你可以使用以下工具：

1. **upload_test_log** - 上传 `go test -json` 输出进行分析
2. **get_analysis_result** - 获取分析结果
3. **terminate_task** - 终止分析任务
4. **get_test_details** - 获取特定测试的详细信息

## 使用流程

1. **生成测试日志**：
   ```bash
   go test -json ./... > test_results.json
   ```

2. **上传并分析**：
   使用 `upload_test_log` 工具上传文件

3. **查看结果**：
   使用返回的 task_id 调用 `get_analysis_result`

4. **查看详细错误**：
   对于失败的测试，使用 `get_test_details` 获取详情

## 注意事项

- Windows 环境下路径中的反斜杠需要转义（使用 `\\`）
- 确保 MCP 客户端有权限访问指定的工作目录
- 测试日志文件必须是 `go test -json` 格式的输出
- 服务器通过标准输入/输出与 MCP 客户端通信