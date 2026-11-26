package handler

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/command"
	"github.com/AlonaBot/AlonaBot/sdk/media"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/menu"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/render"
)

// HelpArguments 保留两级可选选择器，Core 负责命令参数解析。
type HelpArguments struct {
	Plugin  *string `arg:"plugin"`
	Command *string `arg:"command"`
}

// Help 查询当前会话可用命令，选择菜单层级并在图片链路失败时回复同层文本。
func Help(c *sdk.Context, arguments HelpArguments, cache *menu.FontCache, dark bool) error {
	// 1. 只使用 Core 已完成会话准入、权限筛选和本地化的命令。
	entries, err := command.List(c)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return c.Reply(c.T("help.empty"))
	}
	plugins := menu.Group(entries)
	var text string
	var draw func(*render.Fonts, render.Theme) ([]byte, error)
	if arguments.Plugin == nil {
		data := render.IndexData{Title: c.T("help.title"), Subtitle: c.T("help.index_hint")}
		lines := []string{data.Title, data.Subtitle}
		for i, p := range plugins {
			data.Plugins = append(data.Plugins, render.PluginCard{Name: p.Name, Description: p.Description, Count: c.T("help.count", "count", strconv.Itoa(len(p.Commands)))})
			lines = append(lines, helpLine(c, fmt.Sprintf("%d. %s", i+1, p.Name), p.Description))
		}
		text = strings.Join(lines, "\n")
		draw = func(f *render.Fonts, t render.Theme) ([]byte, error) { return render.Index(f, t, data) }
	} else {
		keys := make([][]string, len(plugins))
		labels := make([]string, len(plugins))
		for i, p := range plugins {
			keys[i] = []string{p.ID, p.Name}
			labels[i] = p.Name + " (" + p.ID + ")"
		}
		matches := menu.Match(*arguments.Plugin, keys)
		if len(matches) != 1 {
			return helpMatches(c, *arguments.Plugin, matches, labels)
		}
		plugin := plugins[matches[0]]
		if arguments.Command == nil {
			data := render.DetailData{Title: plugin.Name, Description: plugin.Description, Subtitle: c.T("help.detail_hint", "plugin", strconv.Itoa(matches[0]+1))}
			lines := []string{helpLine(c, plugin.Name, plugin.Description), data.Subtitle}
			for i, entry := range plugin.Commands {
				aliases := menu.Aliases(entry)
				aliasText := ""
				if len(aliases) > 0 {
					aliasText = c.T("help.aliases", "aliases", strings.Join(aliases, ", "))
				}
				data.Commands = append(data.Commands, render.CommandRow{Usage: menu.Usage(entry), Aliases: aliasText, Description: entry.Description})
				lines = append(lines, helpLine(c, fmt.Sprintf("%d. %s", i+1, menu.Usage(entry)), entry.Description))
				if aliasText != "" {
					lines = append(lines, aliasText)
				}
			}
			text = strings.Join(lines, "\n")
			draw = func(f *render.Fonts, t render.Theme) ([]byte, error) { return render.Detail(f, t, data) }
		} else {
			keys := make([][]string, len(plugin.Commands))
			labels := make([]string, len(plugin.Commands))
			for i, entry := range plugin.Commands {
				paths := append([]string{menu.Path(entry.Path)}, menu.Aliases(entry)...)
				for _, path := range paths {
					keys[i] = append(keys[i], path, strings.TrimPrefix(path, "/"))
				}
				labels[i] = helpLine(c, menu.Path(entry.Path), entry.Description)
			}
			matches := menu.Match(*arguments.Command, keys)
			if len(matches) != 1 {
				return helpMatches(c, *arguments.Command, matches, labels)
			}
			entry := plugin.Commands[matches[0]]
			aliases := strings.Join(menu.Aliases(entry), ", ")
			if aliases == "" {
				aliases = c.T("help.none")
			}
			description := entry.Description
			if description == "" {
				description = c.T("help.no_description")
			}
			data := render.CommandData{
				Title: menu.Path(entry.Path), Usage: menu.Usage(entry), Aliases: aliases, Description: description, Plugin: plugin.Name + " (" + plugin.ID + ")",
				UsageLabel: c.T("help.usage_label"), AliasesLabel: c.T("help.aliases_label"), DescriptionLabel: c.T("help.description_label"), PluginLabel: c.T("help.plugin_label"),
			}
			text = strings.Join([]string{data.Title, c.T("help.field", "label", data.UsageLabel, "value", data.Usage), c.T("help.field", "label", data.AliasesLabel, "value", data.Aliases), c.T("help.field", "label", data.DescriptionLabel, "value", data.Description), c.T("help.field", "label", data.PluginLabel, "value", data.Plugin)}, "\n")
			draw = func(f *render.Fonts, t render.Theme) ([]byte, error) { return render.Command(f, t, data) }
		}
	}
	// 2. 字体首次失败由缓存记录警告；绘制、上传与发图失败保留对应原因并回复文本。
	fonts, err := cache.Load(c)
	if err != nil {
		return c.Reply(text)
	}
	theme := render.Light
	if dark {
		theme = render.Dark
	}
	data, err := draw(fonts, theme)
	if err == nil {
		var ref *message.MediaRef
		ref, err = media.Upload(c, bytes.NewReader(data), "image/png")
		if err == nil {
			_, err = c.Send(message.Image(ref.Source()))
		}
	}
	if err == nil {
		return nil
	}
	c.Logger().Warn(fmt.Sprintf("Builtin help could not deliver an image; using the text menu: %v", err))
	return c.Reply(text)
}

func helpLine(c *sdk.Context, name, description string) string {
	if description == "" {
		return c.T("help.line", "names", name)
	}
	return c.T("help.line.description", "names", name, "description", description)
}

func helpMatches(c *sdk.Context, query string, matches []int, labels []string) error {
	if len(matches) == 0 {
		return c.Reply(c.T("help.not_found", "query", query))
	}
	lines := []string{c.T("help.ambiguous", "query", query)}
	for _, i := range matches {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, labels[i]))
	}
	return c.Reply(strings.Join(lines, "\n"))
}
