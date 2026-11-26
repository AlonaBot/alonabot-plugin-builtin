package tools

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/command"
)

// CommandsResponse 返回 Core 按会话和权限过滤的命令、插件元数据及参数。
type CommandsResponse struct {
	Commands []command.Entry `json:"commands"`
}

// CommandsList 保留 Core 提供的路径、参数、全部别名和本地化说明。
func CommandsList(c *sdk.Context, _ ListRequest) (CommandsResponse, error) {
	commands, err := command.List(c)
	return CommandsResponse{Commands: commands}, err
}
