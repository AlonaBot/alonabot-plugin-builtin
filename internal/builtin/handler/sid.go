package handler

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// SID 回复管理员配置白名单时需要复制的会话身份。
func SID(c *sdk.Context) error {
	chatKey := "sid.unknown"
	switch c.ChatType() {
	case message.ChatPrivate:
		chatKey = "sid.private"
	case message.ChatGroup:
		chatKey = "sid.group"
	case message.ChatChannel:
		chatKey = "sid.channel"
	}
	return c.Reply(c.T("sid.identity", "session", c.Session(), "instance", c.Instance(), "user", c.User(), "chat", c.T(chatKey)))
}
