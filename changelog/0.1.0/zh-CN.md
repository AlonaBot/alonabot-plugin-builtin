# 0.1.0

## 新增

- 官方 Builtin 插件 `alonabot-plugin-builtin-official`，完整替代旧 help 插件。
- 纯 Go 绘制三级图片帮助菜单，支持序号、名称、别名、前缀及子串匹配，歧义提示和文本回退。
- `/whoami` 与 `/sid` 身份查询，`/菜单` 帮助别名；开始新对话的 `/new` 由官方 Agent 提供。
- `store_search`、`packages_list`、`packages_install`、`commands_list` 四个 Agent 工具，组件操作要求 Core 管理员身份。
- `dark` 与 `store_source` 配置、中英翻译及 OFL 字体资源。
- 真实 Core 集成测试，以及本地六张菜单样图生成测试。

## 开发依赖

- 当前版本依赖 Core PR #344 的未发布 SDK 接口，开发通过本地 `go.work` 引用 Core 分支。
