// alonabot-plugin-builtin 命令运行官方内置插件。
package main

import (
	"log/slog"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk"
	plugin "github.com/AlonaBot/alonabot-plugin-builtin"
)

func main() {
	if err := sdk.Run(plugin.New()); err != nil {
		slog.Error("Builtin plugin stopped", "error", err)
		os.Exit(1)
	}
}
