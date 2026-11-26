// Package menu 按插件组织 Core 已筛选的命令，解析菜单选择并缓存包内字体。
//
// [Group] 生成稳定序号，[Match] 按精确、前缀和子串顺序匹配，[FontCache] 随插件实例释放。
package menu

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/command"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/render"
)

// Plugin 汇集同一 PluginID 的可用命令及本地化元数据。
type Plugin struct {
	ID, Name, Description string
	Commands              []command.Entry
}

// Group 按显示名排序插件，重名时按 ID 排序，命令按路径排序。
func Group(entries []command.Entry) []Plugin {
	byID := make(map[string]int)
	plugins := make([]Plugin, 0)
	for _, entry := range entries {
		i, found := byID[entry.PluginID]
		if !found {
			i = len(plugins)
			byID[entry.PluginID] = i
			name := entry.PluginName
			if name == "" {
				name = entry.PluginID
			}
			plugins = append(plugins, Plugin{ID: entry.PluginID, Name: name, Description: entry.PluginDescription})
		}
		plugins[i].Commands = append(plugins[i].Commands, entry)
	}
	slices.SortFunc(plugins, func(a, b Plugin) int {
		if order := strings.Compare(a.Name, b.Name); order != 0 {
			return order
		}
		return strings.Compare(a.ID, b.ID)
	})
	for i := range plugins {
		slices.SortFunc(plugins[i].Commands, func(a, b command.Entry) int {
			return strings.Compare(Path(a.Path), Path(b.Path))
		})
	}
	return plugins
}

// Match 返回候选的零基下标；序号按一基解析，各匹配阶段只返回本阶段结果。
func Match(query string, names [][]string) []int {
	if number, err := strconv.Atoi(query); err == nil && number > 0 && number <= len(names) {
		return []int{number - 1}
	}
	query = strings.ToLower(query)
	for _, matches := range []func(string, string) bool{
		func(a, b string) bool { return a == b }, strings.HasPrefix, strings.Contains,
	} {
		var result []int
		for i, aliases := range names {
			for _, name := range aliases {
				if matches(strings.ToLower(name), query) {
					result = append(result, i)
					break
				}
			}
		}
		if len(result) != 0 {
			return result
		}
	}
	return nil
}

// Path 把命令 token 路径写成以斜杠开头的可发送文本。
func Path(tokens []string) string { return "/" + strings.Join(tokens, " ") }

// Usage 拼接命令路径与必填、可选、可变参数提示。
func Usage(entry command.Entry) string {
	parts := []string{Path(entry.Path)}
	for _, param := range entry.Params {
		name := param.Name
		if param.Variadic {
			name += "..."
		}
		if param.Optional {
			parts = append(parts, "["+name+"]")
		} else {
			parts = append(parts, "<"+name+">")
		}
	}
	return strings.Join(parts, " ")
}

// Aliases 返回全部可用别名的命令路径。
func Aliases(entry command.Entry) []string {
	result := make([]string, 0, len(entry.Aliases))
	for _, alias := range entry.Aliases {
		result = append(result, Path(alias))
	}
	return result
}

// FontCache 按插件生命周期缓存字体和首次加载错误；零值可用。
type FontCache struct {
	once  sync.Once
	fonts *render.Fonts
	err   error
}

// Load 首次从 PackageDir 读取字体，加载失败仅记录一次警告。
func (f *FontCache) Load(ctx context.Context) (*render.Fonts, error) {
	f.once.Do(func() { // 同一插件的并发命令共用一次加载结果，失败也缓存。
		dir := sdk.PackageDir(ctx)
		if dir == "" {
			f.err = fmt.Errorf("plugin package directory is empty")
		} else {
			f.fonts, f.err = render.LoadFonts(filepath.Join(dir, "assets", "fonts"))
		}
		if f.err != nil {
			slog.WarnContext(ctx, fmt.Sprintf("Builtin help could not load fonts; using text menus: %v", f.err))
		}
	})
	return f.fonts, f.err
}
