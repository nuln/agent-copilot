# GitHub Copilot LLM — GitHub Copilot LLM 接入

> [English README](README.md)

`github.com/nuln/agent-llm-copilot` — GitHub Copilot LLM 接入插件，适用于 [nuln/agent-core](https://github.com/nuln/agent-core)。

## 概述

| 字段 | 值 |
|------|----|
| **插件类型** | `llm` (LLM 模型) |
| **模块名** | `github.com/nuln/agent-llm-copilot` |
| **核心依赖** | *(仅标准库)* |

## 安装

```bash
go get github.com/nuln/agent-llm-copilot
```

在 `main.go` 中添加副作用导入以激活插件：

```go
import _ "github.com/nuln/agent-llm-copilot"
```

## 配置

通过环境变量或 Web UI 进行配置，完整字段列表请参考源码中的 `RegisterPluginConfigSpec`。

## 开发

```bash
make fmt     # 格式化代码
make lint    # 代码风格检查
make test    # 运行测试
make build   # 编译
```

## 许可证

MIT
