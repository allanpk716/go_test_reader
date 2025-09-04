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

## 示例 3：简化配置（适用于基础 IDE）

某些 IDE 只支持基本的 MCP 配置格式，不支持 `cwd` 和 `env` 参数：

**第一步：编译并设置可执行文件**
```bash
go build -o go_test_reader.exe
```

**第二步：配置 MCP 客户端**
```json
{
  "mcpServers": {
    "go-test-reader": {
      "command": "go_test_reader.exe",
      "args": []
    }
  }
}
```

**第三步：确保可执行文件可访问**

选择以下方式之一：

1. **添加到 PATH 环境变量**（推荐）：
   - 将项目目录添加到系统 PATH
   - 或将 `go_test_reader.exe` 复制到已在 PATH 中的目录

2. **使用完整路径**：
   ```json
   {
     "mcpServers": {
       "go-test-reader": {
         "command": "C:\\WorkSpace\\Go2Hell\\src\\github.com\\allanpk716\\go_test_reader\\go_test_reader.exe",
         "args": []
       }
     }
   }
   ```

**特点：**
- 兼容性最好
- 配置最简单
- 适用于大多数 IDE 和 MCP 客户端

## 注意事项

- Windows 环境下路径中的反斜杠需要转义（使用 `\\`）
- 确保 MCP 客户端有权限访问指定的工作目录
- 测试日志文件必须是 `go test -json` 格式的输出
- 服务器通过标准输入/输出与 MCP 客户端通信
- 对于简化配置，确保可执行文件在 PATH 中或使用完整路径