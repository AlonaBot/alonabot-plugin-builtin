# Browse help and inspect a conversation

[简体中文](usage.zh-CN.md)

Send `/help` or `/菜单` for the plugin overview, `/help 1` for the first plugin, and `/help 1 1` for its first command. The overview has three plugin cards per row. Plugin details list usages, aliases and descriptions; command details show complete information.

## Select by name

Use `/help Builtin` or `/help builtin whoami` to select display and command names. Plugins also accept IDs; commands accept full paths and aliases, such as `/help Builtin /菜单`. Matching tries a valid number, case-insensitive exact names, prefixes, then substrings. Plugin numbers follow display-name order; command numbers follow path order. Numbers may change as available commands change.

A unique match displays an image. Multiple matches return candidates with their original numbers; no match produces a localized message. Core filters commands by conversation and user permissions. Missing fonts, image upload failures or image delivery failures produce the selected level's text menu and log the cause. Library mode without a package directory uses text as well.

## Copy identity details

`/whoami` shows user ID and instance, retaining the old help wording. `/sid` lists session, instance, user and chat type (private, group, channel or unknown) on separate lines. Administrators can copy the session ID into an allowlist. Use the official Agent's `/new` to start a new conversation.

## Configure theme and store

Edit plugin settings in WebUI, or use `data/plugins/alonabot-plugin-builtin-official/config.json` in development mode:

```json
{"dark": true, "store_source": "AlonaBot/plugin-store"}
```

`dark` defaults to `false`. `store_source` is a GitHub `owner/repo`, defaulting to `AlonaBot/plugin-store`. Store search races GitHub Raw and jsDelivr for `store.json` on the `index` branch, accepts the first successful index and cancels the other request. Keywords match repo, id, name, description and tags by case-insensitive substring, optionally combined with a role filter.

## Use the Agent tools

Tool names start with `alonabot-plugin-builtin-official/`:

- `store_search` accepts optional `q` and `role`, and returns `items` with repo, role, id, name, version, description, author, stars and tags. Optional metadata absent from the index may be omitted.
- `packages_list` accepts `{}` and returns `items`, using the SDK's `ID`, `Name`, `Version`, `Role`, `Source` and `Enabled` fields.
- `packages_install` accepts `{"source":"owner/repo"}` and returns `package` and `affected`, preserving Core's partial results and errors. Invocation authorizes initial source trust and same-role replacement. An administrator may need to enable the component in WebUI after installation.
- `commands_list` accepts `{}` and returns `commands`, using the SDK's `PluginID`, `PluginName`, `PluginDescription`, `RouteID`, `Path`, `Params`, `Aliases` and `Description` fields. Parameters contain `Name`, `Optional` and `Variadic`.

Component listing and installation require the parent event's user to be a Core administrator. Tools allow invocation by default; administrators can change policies.

## Languages and resources

Replies and image text use `c.T`, choosing the actor locale, Core default, then English. The English catalog contains messages only; the Chinese catalog also translates command descriptions. Fonts load at runtime from `assets/fonts/`; see [OFL.txt](../assets/fonts/OFL.txt).

Run locally:

```fish
alonabot run . --data-dir ./data
```

See the [README](../README.md) for installation and [development](development.md) for implementation details.
