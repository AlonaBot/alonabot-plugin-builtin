// Package builtin 集中登记官方内置命令、服务与 Agent 工具。
//
// Register 连接 SDK 路由与 handler 包中的命令处理函数。
package builtin

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/service"
	"github.com/AlonaBot/AlonaBot/sdk/tool"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/builtin/handler"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/menu"
	"github.com/AlonaBot/alonabot-plugin-builtin/internal/tools"
)

// Register 在同一个根插件登记全部命令、服务和工具。
func Register(p *sdk.Plugin) {
	fonts := &menu.FontCache{}
	p.Command("help").Alias("菜单").Menu().Description("Show the image help menu").Handle(func(c *sdk.Context, arguments handler.HelpArguments) error {
		return handler.Help(c, arguments, fonts, c.Config().Bool("dark"))
	})
	p.Command("whoami").Menu().Description("Show your user ID and instance").Handle(handler.Whoami)
	p.Command("sid").Menu().Description("Show session, instance, user and chat type").Handle(handler.SID)

	service.Serve(p, service.JSON[tools.SearchRequest, tools.SearchResponse]("/store_search")).Access(service.AccessPublic).
		Tool("Search the component store by an optional case-insensitive keyword q and role (plugin, adapter, agent or webui). Matches repository, ID, name, description and tags. Returns repository and component metadata for installation.", tool.DefaultAllow).
		Effect(tool.ReadOnly).Handle(func(c *sdk.Context, request tools.SearchRequest) (tools.SearchResponse, error) {
		return tools.Search(c, c.Config().String("store_source"), request)
	})
	service.Serve(p, service.JSON[tools.ListRequest, tools.PackagesResponse]("/packages_list")).Access(service.AccessPublic).Permission(sdk.Admin).
		Tool("List installed components with their IDs, names, versions, roles, sources and enabled status. Requires the current user to be a Core administrator.", tool.DefaultAllow).
		Effect(tool.ReadOnly).Handle(tools.PackagesList)
	service.Serve(p, service.JSON[tools.InstallRequest, tools.InstallResponse]("/packages_install")).Access(service.AccessPublic).Permission(sdk.Admin).
		Tool("Install a component from a GitHub owner/repo source. Requires the current user to be a Core administrator and authorizes source trust and same-role replacement. Returns the installation result and affected components. An administrator may need to enable the component in WebUI after installation.", tool.DefaultAllow).
		Effect(tool.Write).Handle(tools.PackagesInstall)
	service.Serve(p, service.JSON[tools.ListRequest, tools.CommandsResponse]("/commands_list")).Access(service.AccessPublic).
		Tool("List commands available to the current user in this conversation, including plugin identity and display name, command paths, required/optional/variadic parameters, aliases and localized descriptions.", tool.DefaultAllow).
		Effect(tool.ReadOnly).Handle(tools.CommandsList)
}
