# 开发官方 Builtin 插件

[English](development.md)

在仓库根目录运行 `go build ./...`，即可编译纯 Go 图片帮助、身份命令与 Agent 工具。Go 版本由 [go.mod](../go.mod) 指定，本机使用 asdf 管理。

## 本地 SDK 前置条件

`go.mod` 保持 `github.com/AlonaBot/AlonaBot v0.3.3`。当前实现依赖 Core [PR #344](https://github.com/AlonaBot/AlonaBot/pull/344) 中的 `sdk.PackageDir`、`sdk/packages`、命令插件元数据及服务会话继承，这些 API 尚未随该 SDK 版本发布。开发环境用被忽略的 `go.work` 同时引用本仓与本地 Core 分支：

```fish
go work init . /path/to/alonabot/.worktrees/builtin-core
```

已有 `go.work` 时用 `go work use` 添加目录。维护者提供的环境已配置工作区，直接复用。依赖使用已批准的 `golang.org/x/image v0.48.0`，无需浏览器或额外运行时。

## 源码职责

| 入口 | 职责 |
| --- | --- |
| `plugin.go` | `New()` 创建根插件，只调用 `internal/builtin.Register(p)` 登记业务 |
| `internal/builtin/register.go` | 集中登记三个命令、四个 JSON 服务与工具权限 |
| `internal/builtin/handler/` | 每命令一文件，处理参数与 `c.T` 回复 |
| `internal/menu/` | 按插件分组、匹配、参数提示及字体懒加载缓存 |
| `internal/render/` | 1080 px 宽度的 PNG，向量圆角卡片与 OpenType 文本排版 |
| `internal/tools/` | 商店竞速、组件操作与当前会话命令查询 |
| `assets/fonts/` | OFL 字体文件及许可，随安装包原样分发 |
| `cmd/alonabot-plugin-builtin/main.go` | `sdk.Run` 进程入口 |
| `plugin_test.go` | 通过公开 `New()` 与 plugintest 驱动真实 Core 和测试 Adapter |

全部业务登记在一个根插件。字体解析后按插件实例缓存；每次渲染创建独立 `font.Face`。字体失败也缓存。包目录为空或资源缺失会记录一次英文 warn，并使用文本菜单。`render.ParseFonts` 接受调用方的字体字节，测试也可以直接传入字体；运行时由 `sdk.PackageDir(ctx)` 定位 `assets/fonts/`，字体随资源分发。

## SDK 调用契约

帮助数据来自 `command.List(c)`，保留 Core 的准入、权限和本地化结果。图片通过 `media.Upload(c, bytes.NewReader(png), "image/png")` 上传，再发送 `message.Image(ref.Source())`。渲染、上传或发送失败记录原因并回复同层文本。

SDK 当前提供的 HTTP 能力为入站路由。商店使用 `http.DefaultClient` 和继承父请求的 context，同时请求 GitHub Raw 与 jsDelivr，首个成功索引取消另一请求；两端均失败时保留全部错误。请求沿用调用期限。

四个工具使用 `service.JSON[In, Out]("/<名称>")` 与 `service.Serve` 注册，声明公开访问、默认允许及 ReadOnly/Write 副作用。组件列表和安装额外声明 `sdk.Admin`，Core 根据原消息 Actor 执行权限检查。组件安装保留部分结果、affected 和完整错误。

JSON 输入输出必须是包级具名非泛型类型。schema ID 来自完整包路径与类型名；请求与响应同时校验 ID 和版本，同形结构体也各有身份。Agent 工具使用 `tool.Invoke`，由 Core 根据已登记的服务 schema 转发参数。Builtin 的业务代码无需导入 Agent 契约。开始新对话的命令归官方 Agent。

## 验证改动

```fish
gofmt -w plugin.go plugin_test.go internal cmd
gofmt -l .
go build ./...
go vet ./...
go test ./...
go test -race ./...
alonabot check .
alonabot test .
```

集成测试验证 `/whoami`、私聊与群聊 `/sid`、三级菜单、序号与名称匹配、歧义与未命中、图片发送失败的文本回退，以及通过真实工具调用链执行 `commands_list`。测试 Adapter 替代外部平台，保留真实 Core 与 SDK。沙箱禁止监听 loopback 端口时，记录原始错误，由具备端口权限的本机环境复测。

本地渲染测试位于已忽略的 `internal/render/render_local_test.go`，读取真实字体，生成浅色与深色的一级、二级、三级共六张 PNG，检查宽度 1080 px、字符换行和省略号。仅运行该测试：

```fish
go test ./internal/render -run TestRenderMenusLocal -v
```

图片位于 `tmp/render/{light,dark}-{index,detail,command}.png`，人工检查文字、边框和卡片布局。本地商店测试使用内存 HTTP Transport 验证双源竞速取消、过滤、无效输入和双源失败，运行 `go test ./internal/tools`。本地测试使用 `*_local_test.go`，只保留真实入口集成测试作为提交内容。

CI 使用五平台 pure-Go 构建。手动执行：

```fish
for target in linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64
    set -l parts (string split / $target)
    env GOOS=$parts[1] GOARCH=$parts[2] CGO_ENABLED=0 go build ./...
end
```

修改中英文文档与词典时保持同步。依赖与字体例外见[编码规范末尾](coding-standards.zh-CN.md#本仓差异)，贡献流程见[贡献指南](../CONTRIBUTING.zh-CN.md)。
