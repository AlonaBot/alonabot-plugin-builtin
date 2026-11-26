// Package tools 实现官方内置插件公开给 AI Agent 的商店、组件与命令查询服务。
//
// JSON 服务由 builtin.Register 集中登记，权限与父事件身份由 SDK 和 Core 执行。
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

// SearchRequest 指定可选的大小写无关关键词与组件角色。
type SearchRequest struct {
	Q    string `json:"q,omitempty"`
	Role string `json:"role,omitempty"` // plugin、adapter、agent 或 webui；空串包含全部角色。
}

// StoreEntry 保留商店索引向工具公开的字段。
type StoreEntry struct {
	Repo        string   `json:"repo"`
	Role        string   `json:"role"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Stars       int      `json:"stars,omitempty"`
	Tags        []string `json:"tags"`
}

// SearchResponse 返回按原索引顺序筛选的组件。
type SearchResponse struct {
	Items []StoreEntry `json:"items"`
}

var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+$`)

// Search 并发读取两个商店镜像，首个成功响应取消另一请求，再筛选完整索引。
func Search(ctx context.Context, source string, request SearchRequest) (SearchResponse, error) {
	// 1. 校验配置和模型输入，仓库字段只允许 GitHub owner/repo。
	_, repository, _ := strings.Cut(source, "/")
	if !repositoryPattern.MatchString(source) || repository == "." || repository == ".." {
		return SearchResponse{}, errcode.InvalidArgument.New("store_source must be a GitHub owner/repo")
	}
	switch request.Role {
	case "", "plugin", "adapter", "agent", "webui":
	default:
		return SearchResponse{}, errcode.InvalidArgument.New("role must be plugin, adapter, agent or webui")
	}
	// 2. 每个请求向缓冲 channel 交付一次结果，取消后交付也不会留下阻塞的 goroutine。
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type response struct {
		items []StoreEntry
		err   error
	}
	urls := []string{"https://raw.githubusercontent.com/" + source + "/index/store.json", "https://cdn.jsdelivr.net/gh/" + source + "@index/store.json"}
	responses := make(chan response, len(urls))
	client := http.DefaultClient
	for _, address := range urls {
		go func() { items, err := fetchStore(ctx, client, address); responses <- response{items, err} }()
	}
	var failures []error
	for range urls {
		select { // 父调用取消时立即返回；只有成功解码的索引才赢得镜像竞速。
		case <-ctx.Done():
			return SearchResponse{}, errors.Join(append(failures, ctx.Err())...)
		case result := <-responses:
			if result.err != nil {
				failures = append(failures, result.err)
				continue
			}
			cancel()
			items := make([]StoreEntry, 0)
			query := strings.ToLower(request.Q)
			for _, item := range result.items {
				if request.Role != "" && item.Role != request.Role {
					continue
				}
				fields := []string{item.Repo, item.ID, item.Name, item.Description}
				fields = append(fields, item.Tags...)
				for _, field := range fields {
					if strings.Contains(strings.ToLower(field), query) {
						items = append(items, item)
						break
					}
				}
			}
			return SearchResponse{Items: items}, nil
		}
	}
	return SearchResponse{}, errors.Join(failures...)
}

func fetchStore(ctx context.Context, client *http.Client, address string) (items []StoreEntry, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, fmt.Errorf("create store request %s: %w", address, err)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch store %s: %w", address, err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close store response %s: %w", address, closeErr))
		}
	}()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read store %s: %w", address, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch store %s: %s: %s", address, response.Status, string(body))
	}
	var index struct {
		Items *[]StoreEntry `json:"items"`
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("decode store %s: %w", address, err)
	}
	if index.Items == nil {
		return nil, fmt.Errorf("decode store %s: missing items array", address)
	}
	return *index.Items, nil
}
