# AlonaBot Official Builtin Plugin

[简体中文](README.zh-CN.md)

Send `/help` to open a PicMenu-style, three-level image menu. Use `/whoami` or `/sid` to inspect user and conversation identity. Builtin fully replaces the old help plugin and gives AI Agents tools for store search, installed components, component installation and available commands. The official Agent provides `/new` to start a new conversation.

The plugin ID is `alonabot-plugin-builtin-official`; its Go module is `github.com/AlonaBot/alonabot-plugin-builtin`. Images are drawn in pure Go with light and dark themes.

## Install

```fish
alonabot install --source AlonaBot/alonabot-plugin-builtin --yes
```

An administrator enables the plugin in WebUI and disables the old help plugin to avoid duplicate commands. This development version uses unreleased SDK APIs from Core [PR #344](https://github.com/AlonaBot/AlonaBot/pull/344), and requires a Core containing those APIs. See [development](docs/development.md) for the local SDK setup.

## Commands

| Command | Alias | Purpose |
| --- | --- | --- |
| `/help [plugin] [command]` | `/菜单` | Overview → plugin details → command details; accepts numbers, names and partial matches |
| `/whoami` | — | Show user ID and instance |
| `/sid` | — | Show session ID, instance, user ID and localized chat type for allowlist configuration |

Core filters the list by conversation admission and permissions. Font loading, image upload or image delivery failures produce a text menu for the selected level. See [usage](docs/usage.md).

## Agent tools

Full tool names are `alonabot-plugin-builtin-official/<name>`, with an allow policy by default.

| Name | Input | Effect | Permissions and result |
| --- | --- | --- | --- |
| `store_search` | Optional `q`, `role` | ReadOnly | Search the store; role is plugin, adapter, agent or webui |
| `packages_list` | `{}` | ReadOnly | Core administrator; component identities and enabled status |
| `packages_install` | `source: "owner/repo"` | Write | Core administrator; installation result and affected components; an administrator may need to enable the component in WebUI |
| `commands_list` | `{}` | ReadOnly | Available plugin identities, command paths, parameters, aliases and descriptions in this conversation |

## Configuration

| Field | Type | Default | Purpose |
| --- | --- | --- | --- |
| `dark` | bool | `false` | Use the dark help menu theme |
| `store_source` | string | `AlonaBot/plugin-store` | GitHub repository containing the store index |

Fonts ship in `assets/fonts/` under the [SIL Open Font License](assets/fonts/OFL.txt) and are read from the package directory at runtime.

See [development](docs/development.md), [coding standards](docs/coding-standards.md) and [releasing](docs/releasing.md).
