// Package handler 处理内置命令的输入与本地化回复，每个命令单独登记处理函数。
package handler

import "github.com/AlonaBot/AlonaBot/sdk"

// Whoami 回复触发者的用户 ID 与实例，保留旧 help 的文案。
func Whoami(c *sdk.Context) error {
	if c.User() == "" {
		return c.Reply(c.T("help.whoami.no_actor"))
	}
	return c.Reply(c.T("help.whoami", "user", c.User(), "instance", c.Instance()))
}
