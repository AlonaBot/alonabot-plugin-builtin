# AlonaBot 官方 Builtin 插件

[English](README.md)

发送 `/help` 查看 PicMenu 风格的三级图片帮助菜单；发送 `/whoami` 或 `/sid` 查询用户与会话身份。Builtin 完全替代旧 help 插件，并向 AI Agent 提供商店搜索、已装组件列表、安装组件和命令列表工具。开始新对话的 `/new` 由官方 Agent 提供。

插件 ID 为 `alonabot-plugin-builtin-official`，模块为 `github.com/AlonaBot/alonabot-plugin-builtin`。图片由纯 Go 绘制，支持浅色与深色主题。

## 安装

```fish
alonabot install --source AlonaBot/alonabot-plugin-builtin --yes
```

安装后由管理员在 WebUI 启用，并停用旧 help 插件，避免重复命令。当前开发版本使用 Core [PR #344](https://github.com/AlonaBot/AlonaBot/pull/344) 的未发布 SDK 接口；运行时需要包含这些接口的 Core。开发接入方式见[开发指南](docs/development.zh-CN.md)。

## 命令

| 命令 | 别名 | 用途 |
| --- | --- | --- |
| `/help [plugin] [command]` | `/菜单` | 总览 → 插件详情 → 命令详情，支持序号、名称与模糊匹配 |
| `/whoami` | — | 用户 ID 与实例 |
| `/sid` | — | 会话 ID、实例、用户 ID 与本地化会话类型，方便配置白名单 |

列表由 Core 按当前会话准入与权限筛选。字体加载、图片上传或发送失败时，帮助菜单回复对应层级的文本。详情见[使用指南](docs/usage.zh-CN.md)。

## Agent 工具

完整工具名称为 `alonabot-plugin-builtin-official/<名称>`，默认策略为允许。

| 名称 | 输入 | 副作用 | 权限与结果 |
| --- | --- | --- | --- |
| `store_search` | `q?`、`role?` | ReadOnly | 搜索商店；role 为 plugin、adapter、agent 或 webui |
| `packages_list` | `{}` | ReadOnly | Core 管理员；返回组件身份与启用状态 |
| `packages_install` | `source: "owner/repo"` | Write | Core 管理员；返回安装结果和 affected，安装后可能需要管理员在 WebUI 启用 |
| `commands_list` | `{}` | ReadOnly | 返回当前会话可用命令的插件、路径、参数、别名和说明 |

## 配置

| 字段 | 类型 | 默认值 | 用途 |
| --- | --- | --- | --- |
| `dark` | bool | `false` | 深色帮助菜单 |
| `store_source` | string | `AlonaBot/plugin-store` | 商店索引所在的 GitHub 仓库 |

字体随 `assets/fonts/` 分发，遵循 [SIL Open Font License](assets/fonts/OFL.txt)，运行时按包目录读取。

开发与验证见[开发指南](docs/development.zh-CN.md)，规范见[编码规范](docs/coding-standards.zh-CN.md)，发布见[发布指南](docs/releasing.zh-CN.md)。
