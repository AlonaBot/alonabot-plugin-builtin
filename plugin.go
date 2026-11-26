// Package builtin 提供示例插件的公开构造入口。
package builtin

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	implementation "github.com/AlonaBot/alonabot-plugin-builtin/internal/builtin"
)

// New 创建独立插件并登记全部命令。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()
	implementation.Register(p)
	return p
}
