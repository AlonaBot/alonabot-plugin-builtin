# 0.1.0

## Added

- Official Builtin plugin `alonabot-plugin-builtin-official`, fully replacing the old help plugin.
- Three-level image help drawn in pure Go, with numeric, name, alias, prefix and substring matching, ambiguity messages and text fallback.
- `/whoami` and `/sid` identity queries and the `/菜单` help alias. The official Agent provides `/new` for new conversations.
- Four Agent tools: `store_search`, `packages_list`, `packages_install` and `commands_list`. Component operations require a Core administrator.
- `dark` and `store_source` configuration, English/Chinese catalogs and OFL font resources.
- Real Core integration tests and a local test that generates six sample menus.

## Development dependency

- This version currently requires unreleased SDK APIs from Core PR #344, referenced through a local `go.work` during development.
