# 浏览帮助与查询会话

[English](usage.md)

发送 `/help` 或 `/菜单` 查看插件总览，再发送 `/help 1` 打开第一个插件，最后用 `/help 1 1` 查看它的第一个命令。一级菜单每行显示三张插件卡片；二级列出用法、别名与说明；三级展示命令的完整信息。

## 按名称选择

`/help Builtin` 与 `/help builtin whoami` 使用显示名和命令名选择。插件也支持 ID；命令支持完整路径与别名，例如 `/help Builtin /菜单`。匹配顺序为有效序号、忽略大小写的精确名称、前缀、子串。每个插件的序号按显示名排序，命令序号按路径排序；可见命令变化时序号也可能变化。

唯一匹配直接显示图片；多个匹配回复带原序号的候选列表；未命中回复本地化提示。Core 按当前会话和用户权限筛选命令。字体不可用或图片上传、发送失败时，自动回复所选层级的文本，并在日志中保留失败原因。库模式缺少包目录时也使用文本。

## 复制身份信息

`/whoami` 显示用户 ID 与实例，沿用旧 help 的文案。`/sid` 逐行显示会话、实例、用户和会话类型（私聊、群聊、频道或未知）。管理员可复制会话 ID 到白名单配置。开始新对话请使用官方 Agent 的 `/new`。

## 配置主题与商店

在 WebUI 编辑插件配置；开发模式可编辑 `data/plugins/alonabot-plugin-builtin-official/config.json`：

```json
{"dark": true, "store_source": "AlonaBot/plugin-store"}
```

`dark` 默认 `false`。`store_source` 为 GitHub `owner/repo`，默认 `AlonaBot/plugin-store`。商店搜索同时读取 GitHub Raw 与 jsDelivr 的 `index` 分支 `store.json`，采用首个成功索引并取消另一请求。关键词对 repo、id、name、description、tags 做忽略大小写的子串匹配，可叠加角色筛选。

## 让 Agent 使用工具

工具前缀为 `alonabot-plugin-builtin-official/`：

- `store_search` 接收可选 `q` 与 `role`，返回 `items`，包含 repo、role、id、name、version、description、author、stars、tags；索引缺省的可选字段可省略。
- `packages_list` 接收 `{}`，返回 `items`。每项采用 SDK 的 `ID`、`Name`、`Version`、`Role`、`Source`、`Enabled` 字段。
- `packages_install` 接收 `{"source":"owner/repo"}`，返回 `package` 与 `affected`，保留 Core 的部分成功结果和错误。调用授权首次来源信任及同类替换，安装后可能需要管理员在 WebUI 启用。
- `commands_list` 接收 `{}`，返回 `commands`。每项采用 SDK 的 `PluginID`、`PluginName`、`PluginDescription`、`RouteID`、`Path`、`Params`、`Aliases`、`Description`；参数含 `Name`、`Optional`、`Variadic`。

组件列表和安装要求父事件的用户具有 Core 管理员身份。工具默认允许，管理端仍可调整调用策略。

## 语言与资源

回复和图片文字通过 `c.T`，按 Actor 语言、Core 默认语言、英文回退选择。英文词典只含 messages；中文词典同时翻译命令说明。字体位于 `assets/fonts/`，运行时加载，详细许可见 [OFL.txt](../assets/fonts/OFL.txt)。

本地运行：

```fish
alonabot run . --data-dir ./data
```

安装命令见 [README](../README.zh-CN.md)，源码说明见[开发指南](development.zh-CN.md)。
