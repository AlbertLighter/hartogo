# 项目概述

本项目 `hartogo` 是一个使用 Go 语言编写的命令行工具。其主要功能是将 HTTP Archive (HAR) 文件转换为可执行的 Go 代码。

对于给定 HAR 文件中的每个条目，该工具都会生成一个独立的 `.go` 文件。每个生成的文件都包含一个 Go 函数，该函数使用 `resty.dev/v3` 库来复现原始的 HTTP 请求。该工具还会尝试为 JSON 格式的请求和响应体自动生成 Go 结构体，从而简化数据处理。

核心逻辑位于 `internal/converter` 包中，该包负责处理 HAR 解析、JSON 到 Go 结构体的转换，以及通过 Go 模板 (`internal/converter/templates/resty.tmpl`) 生成代码。命令行工具的主入口点位于 `cmd/hartogo/main.go`。

# 构建与运行

## 构建工具

要构建 `hartogo` 可执行文件，请在项目根目录运行以下命令：

```sh
go build ./cmd/hartogo
```

## 运行工具

要使用该工具，您需要提供一个 HAR 文件的路径。生成的 Go 文件将被放置在一个以输入文件名命名的新目录中（例如，`example.har` 会生成 `example_req` 目录）。

```sh
# ./hartogo -input <har_file_path>
./hartogo -input example/example.har
```

您可以使用 `-output-dir` 标志指定一个不同的输出目录：

```sh
# ./hartogo -input <har_file_path> -output-dir <your_directory>
./hartogo -input example/example.har -output-dir my_generated_requests
```

生成的 Go 文件将属于 `requests` 包。

# 开发约定

*   **代码风格**: 代码使用标准的 `go fmt` 进行格式化。
*   **依赖管理**: 项目依赖通过 Go Modules (`go.mod` 和 `go.sum`) 进行管理。主要的外部依赖是用于发起 HTTP 请求的 `resty.dev/v3`。
*   **项目结构**: 项目遵循标准的 Go 布局，主应用位于 `cmd` 目录，可复用逻辑位于 `internal` 目录。
*   **代码生成**: 该工具严重依赖 Go 的 `text/template` 包从模板文件生成代码。这使得修改生成的请求文件的输出格式变得容易。
