# hartogo

`hartogo` 是一个使用 Go 语言编写的命令行工具，它可以将 HTTP Archive (HAR) 文件转换为可执行的 Go 代码。此外，它还提供了一个实用功能，可以直接从 JSON 文件生成 Go 结构体。

## 概述

本工具旨在简化为测试、调试或 API 客户端实现而重建 HTTP 请求的流程。对于 HAR 文件中的每个条目，它都会生成一个独立的 `.go` 文件，其中包含一个使用 `resty.dev/v3` 库来复现原始 HTTP 请求的 Go 函数。

此外，它能够检查 JSON 内容（无论是来自 HAR 条目还是独立的 `.json` 文件），并尝试自动生成相应的 Go 结构体，从而简化基于 JSON 的 API 的数据处理。

## 功能特性

- **HAR 转 Go 代码：** 将 HAR 文件中的每个请求转换为一个立即可用的 Go 函数。
- **JSON 转 Go 结构体：** 从 JSON 数据自动生成 Go 结构体，能够处理嵌套对象和数组。
- **`resty` 集成：** 生成的代码使用流行且功能强大的 `resty.dev/v3` 客户端来发起 HTTP 请求。
- **自定义输出：** 允许为生成的文件指定一个自定义的输出目录。

## 安装

您可以通过两种方式安装 `hartogo`：

### 1. 使用 `go install` (推荐)

您可以运行以下命令来安装 `hartogo` 到您的 `GOPATH` 中。这允许您在系统的任何路径下直接运行 `hartogo` 命令。

```sh
go install github.com/AlbertLighter/hartogo/cmd/hartogo@latest
```

请确保您的 `GOPATH/bin` 目录已经添加到了系统的 `PATH` 环境变量中。

### 2. 本地构建

如果您想在本地编译，请克隆本仓库，然后在项目根目录运行以下命令：

```sh
go build ./cmd/hartogo
```

这会在当前目录下生成一个 `hartogo` 可执行文件。

## 使用方法

该工具支持两种模式：转换完整的 HAR 文件，或从单个 JSON 文件生成结构体。

### 转换 HAR 文件

使用 `-input` 标志提供 HAR 文件的路径。生成的 Go 文件将被放置在一个以输入文件名命名的新目录中（例如，`example.har` 会创建 `example_req/` 目录）。

```sh
# 基本用法
./hartogo -input example/example.har

# 指定自定义输出目录
./hartogo -input example/example.har -output-dir my_generated_requests
```

生成的 Go 文件将属于 `requests` 包。

### 从 JSON 生成结构体

如果您提供一个 `.json` 文件作为输入，该工具将生成一个包含相应结构体的 Go 文件。这对于从 JSON 示例快速创建数据模型非常有用。

```sh
# 从 JSON 文件生成 Go 结构体
./hartogo -input /path/to/your/data.json -output-dir /path/to/your/models
```

如果 JSON 文件名以数字开头（例如 `1.json`），生成的结构体名称将以 `Json` 为前缀（例如 `Json1`），以确保它是一个有效的 Go 标识符。

## 开发

- **代码风格**：代码使用标准的 `go fmt` 进行格式化。
- **依赖管理**：项目依赖通过 Go Modules (`go.mod`) 进行管理。主要的外部依赖是 `resty.dev/v3`。
- **项目结构**：项目遵循标准的 Go 布局，主应用位于 `cmd` 目录，可复用逻辑位于 `internal` 目录。
- **代码生成**：该工具严重依赖 Go 的 `text/template` 包从模板生成代码，这使得修改生成的请求文件的输出格式变得容易。