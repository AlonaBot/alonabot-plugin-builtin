# Develop the official Builtin plugin

[简体中文](development.zh-CN.md)

Run `go build ./...` from the repository root to compile the pure Go image menu, identity commands and Agent tools. Use the Go version declared in [go.mod](../go.mod), managed locally by asdf.

## Local SDK prerequisite

`go.mod` retains `github.com/AlonaBot/AlonaBot v0.3.3`. The implementation currently uses `sdk.PackageDir`, `sdk/packages`, command plugin metadata and inherited service conversation context from Core [PR #344](https://github.com/AlonaBot/AlonaBot/pull/344). These APIs have not shipped in that SDK version. An ignored `go.work` references both this repository and the local Core branch:

```fish
go work init . /path/to/alonabot/.worktrees/builtin-core
```

Use `go work use` to add directories to an existing workspace. The maintainer's supplied environment already has this workspace configured. The approved `golang.org/x/image v0.48.0` dependency provides drawing support without a browser or extra runtime.

## Source ownership

| Entry | Responsibility |
| --- | --- |
| `plugin.go` | `New()` creates the root plugin and delegates business registration to `internal/builtin.Register(p)` |
| `internal/builtin/register.go` | Register three commands and four JSON services, tools and permissions |
| `internal/builtin/handler/` | One file per command; arguments and `c.T` replies |
| `internal/menu/` | Grouping, matching, parameter hints and lazy font caching |
| `internal/render/` | 1080 px PNGs, vector rounded cards and OpenType text layout |
| `internal/tools/` | Store mirror race, package operations and available-command queries |
| `assets/fonts/` | OFL fonts and license, copied into the installation package |
| `cmd/alonabot-plugin-builtin/main.go` | `sdk.Run` process entry |
| `plugin_test.go` | Real Core and test Adapter through public `New()` and plugintest |

All business routes use one root plugin. Parsed fonts are cached per plugin instance; each render owns its `font.Face` values. Failed loads are cached as well. A missing package directory or font logs one English warning and uses text menus. `render.ParseFonts` accepts caller-provided font bytes for direct tests. Runtime loading resolves `assets/fonts/` through `sdk.PackageDir(ctx)`; font files ship as resources.

## SDK contracts

Help uses `command.List(c)` with Core admission, permission filtering and localization. Upload PNGs with `media.Upload(c, bytes.NewReader(png), "image/png")`, then send `message.Image(ref.Source())`. Rendering, upload or delivery errors log the cause and produce the same menu level as text.

The SDK currently exposes inbound HTTP routes. Store requests use `http.DefaultClient` and the parent's context, race GitHub Raw against jsDelivr, and cancel the other request when the first valid index arrives. If both fail, all errors are retained. Requests inherit the call deadline.

Each tool registers with `service.JSON[In, Out]("/<name>")` and `service.Serve`, with public access, a default allow policy and ReadOnly/Write effects. Component listing and installation additionally require `sdk.Admin`; Core checks the original message actor. Installation preserves partial results, affected components and the full error.

JSON inputs and outputs must be package-level named, non-generic types. Schema IDs include the full package path and type name. Requests and responses check IDs and versions; structurally identical types still have separate identities. Agents call tools with `tool.Invoke`, letting Core forward arguments using the registered schema. Builtin business code needs no Agent contract import. The official Agent owns new-conversation commands.

## Verify changes

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

Integration tests cover `/whoami`, private/group `/sid`, all three menu levels, numeric/name matching, ambiguous and missing matches, text fallback after image rejection, and `commands_list` through the real tool invocation chain. The test Adapter replaces the external platform while Core and SDK remain real. If sandbox permissions prohibit loopback listeners, record the exact error and rerun on a local machine with port access.

The ignored `internal/render/render_local_test.go` loads real fonts and generates six PNGs: light/dark overview, plugin and command menus. It checks 1080 px width, character wrapping and ellipsis. Run it alone with:

```fish
go test ./internal/render -run TestRenderMenusLocal -v
```

Inspect `tmp/render/{light,dark}-{index,detail,command}.png` for text, borders and layout. Local store tests use an in-memory HTTP Transport to verify mirror cancellation, filtering, invalid input and dual failures; run `go test ./internal/tools`. Local tests use `*_local_test.go`; only integration tests through real entry points are intended for submission.

CI builds pure Go for five platforms. To run these builds manually:

```fish
for target in linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64
    set -l parts (string split / $target)
    env GOOS=$parts[1] GOARCH=$parts[2] CGO_ENABLED=0 go build ./...
end
```

Keep bilingual documentation and catalogs synchronized. See the [repository differences](coding-standards.md#repository-differences) for drawing/font exceptions and the [contribution guide](../CONTRIBUTING.md).
