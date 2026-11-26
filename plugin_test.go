package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/tool"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/tools"
)

// TestCommands 经公开 New、真实 Core 与测试 Adapter 验证身份、菜单和工具调用链。
func TestCommands(t *testing.T) {
	for _, scenario := range []struct {
		name         string
		config       string
		textFallback bool
	}{
		{name: "light", config: `{}`},
		{name: "dark_text_fallback", config: `{"dark":true}`, textFallback: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			p := New()
			p.Command("sample").Description("Sample command").Handle(func(*sdk.Context) error { return nil })
			p.Command("sample-extra").Description("Another sample").Handle(func(*sdk.Context) error { return nil })
			p.Command("test-commands").Handle(func(c *sdk.Context) error {
				result, err := tool.Invoke(c, "alonabot-plugin-builtin-official/commands_list", json.RawMessage(`{}`))
				if err != nil {
					return err
				}
				return c.Reply(string(result))
			})
			options := []plugintest.Option{plugintest.Plugin(p), plugintest.Config(p, scenario.config)}
			if scenario.textFallback {
				options = append(options, plugintest.Operations(plugintest.Behavior{Send: func(_ context.Context, call plugintest.Call) (*message.SendResult, *errcode.Error) {
					for _, segment := range call.Chain {
						if _, ok := segment.(message.ImageSegment); ok {
							return nil, errcode.Unsupported.New("test adapter rejects images")
						}
					}
					return call.SendResult, nil
				}}))
			}
			h := plugintest.Start(t, options...)
			user := plugintest.User{ID: "example-user"}
			chat := plugintest.Private(user.ID)
			h.Text(chat, user, "/whoami").RequireReply(t, "User: "+message.UID(plugintest.AdapterID, user.ID)+"\nInstance: "+plugintest.InstanceID)
			h.Text(chat, user, "/sid").RequireReply(t, fmt.Sprintf("Session: %s\nInstance: %s\nUser: %s\nChat type: Private", message.SID(plugintest.AdapterID, "private:"+user.ID, ""), plugintest.InstanceID, message.UID(plugintest.AdapterID, user.ID)))
			h.Text(plugintest.Group("room"), user, "/sid").RequireReply(t, fmt.Sprintf("Session: %s\nInstance: %s\nUser: %s\nChat type: Group", message.SID(plugintest.AdapterID, "group:room", ""), plugintest.InstanceID, message.UID(plugintest.AdapterID, user.ID)))
			for _, test := range []struct{ command, contains string }{
				{"/help", "Help Menu"}, {"/菜单", "Help Menu"}, {"/help 1", "Builtin"}, {"/help BUILTIN", "Builtin"},
				{"/help alonabot-plugin-builtin-official", "Builtin"}, {"/help buil", "Builtin"}, {"/help uiltin", "Builtin"},
				{"/help 1 1", "/help"}, {"/help 1 help", "/help"}, {"/help 1 /菜单", "/help"},
				{"/help 1 WHOAMI", "/whoami"}, {"/help 1 who", "/whoami"}, {"/help 1 hoam", "/whoami"}, {"/help 1 sample", "/sample"},
			} {
				t.Run(test.command, func(t *testing.T) {
					result := h.Text(chat, user, test.command)
					for _, outcome := range result.Outcomes {
						if outcome.Err != nil {
							t.Fatalf("handler failed: %v", outcome.Err)
						}
					}
					sends := result.Sends
					if scenario.textFallback {
						if len(sends) != 2 || sends[0].Err == nil || sends[0].Err.Code != errcode.Unsupported {
							t.Fatalf("sends = %+v, want a rejected image followed by text", sends)
						}
						sends = sends[1:]
					}
					if len(sends) != 1 {
						t.Fatalf("sends = %+v, want one menu", result.Sends)
					}
					sent := sends[0]
					if sent.Err != nil {
						t.Fatalf("menu send failed: %v", sent.Err)
					}
					image := false
					for _, segment := range sent.Chain {
						if _, ok := segment.(message.ImageSegment); ok {
							image = true
						}
					}
					if image {
						if scenario.textFallback {
							t.Fatal("expected text after image rejection")
						}
						if len(sent.Media) != 1 {
							t.Fatalf("media = %+v, want one PNG", sent.Media)
						}
						decoded, format, err := sent.Media[0].Image()
						if err != nil {
							t.Fatal(err)
						}
						if format != "png" || decoded.Bounds().Dx() != 1080 {
							t.Fatalf("menu = %s %v, want PNG width 1080", format, decoded.Bounds())
						}
					} else if !strings.Contains(sent.Text(), test.contains) {
						t.Fatalf("reply = %q, want %q", sent.Text(), test.contains)
					}
				})
			}
			h.Text(chat, user, "/help missing").RequireReply(t, "No match found for “missing”.")
			h.Text(chat, user, "/help 1 missing").RequireReply(t, "No match found for “missing”.")
			h.Text(chat, user, "/help 1 sam").RequireReply(t, "Multiple matches for “sam”; choose a number or a more specific name:\n2. /sample — Sample command\n3. /sample-extra — Another sample")
			result := h.Text(chat, user, "/test-commands")
			for _, outcome := range result.Outcomes {
				if outcome.Err != nil {
					t.Fatalf("tool call failed: %v", outcome.Err)
				}
			}
			if len(result.Sends) != 1 {
				t.Fatalf("tool sends = %+v, want one", result.Sends)
			}
			var response tools.CommandsResponse
			if err := json.Unmarshal([]byte(result.Sends[0].Text()), &response); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, entry := range response.Commands {
				path := strings.Join(entry.Path, " ")
				if path == "new" || path == "reset" || path == "hello" {
					t.Fatalf("unexpected command %q", path)
				}
				if path != "help" {
					continue
				}
				found = true
				if entry.PluginID != "alonabot-plugin-builtin-official" || entry.PluginName != "Builtin" || entry.Description != "Show the image help menu" || len(entry.Params) != 2 || !entry.Params[0].Optional || !entry.Params[1].Optional || len(entry.Aliases) != 1 || strings.Join(entry.Aliases[0], " ") != "菜单" {
					t.Fatalf("help entry = %+v", entry)
				}
			}
			if !found {
				t.Fatalf("help missing from commands tool: %+v", response)
			}
		})
	}
}
