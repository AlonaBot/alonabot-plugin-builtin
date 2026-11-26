package tools

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/packages"
)

// ListRequest 提供无参数查询服务的具名 JSON 输入。
type ListRequest struct{}

// PackagesResponse 返回 Core 报告的已装组件身份和启用状态。
type PackagesResponse struct {
	Items []packages.Info `json:"items"`
}

// PackagesList 使用原始父事件的管理员身份查询组件。
func PackagesList(c *sdk.Context, _ ListRequest) (PackagesResponse, error) {
	items, err := packages.List(c)
	return PackagesResponse{Items: items}, err
}

// InstallRequest 指定 GitHub owner/repo 安装来源。
type InstallRequest struct {
	Source string `json:"source"`
}

// InstallResponse 保留安装结果、受影响对象以及可能同时返回的业务错误。
type InstallResponse struct {
	Package  packages.Info `json:"package"`
	Affected []string      `json:"affected"`
}

// PackagesInstall 把来源交由 Core 校验并执行安装，原样保留部分结果。
func PackagesInstall(c *sdk.Context, request InstallRequest) (InstallResponse, error) {
	installed, affected, err := packages.Install(c, request.Source)
	return InstallResponse{Package: installed, Affected: affected}, err
}
