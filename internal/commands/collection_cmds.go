package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/client"
	"github.com/forechoandlook/bilibili-cli/internal/formatter"
)

func AddCollectionCommands(rootCmd *cobra.Command) {
	var page int
	var maxCount int

	favoritesCmd := &cobra.Command{
		Use:   "favorites [ID]",
		Short: "查看收藏夹或收藏夹内容",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(false, outputJSON, outputYAML)
			api := client.NewClient()

			if page <= 0 {
				formatter.ExitError("invalid_input", "page 必须大于 0", format)
			}

			if len(args) == 0 {
				// Get self info to get mid
				infoData, err := api.Call(cmd.Context(), "获取当前登录用户信息", api.R(cmd.Context(), cred), "GET", "https://api.bilibili.com/x/space/myinfo")
				if err != nil {
					formatter.ExitError("upstream_error", "获取用户信息失败: "+err.Error(), format)
				}
				var info map[string]interface{}
				json.Unmarshal(infoData, &info)

				data, err := api.Call(cmd.Context(), "获取收藏夹列表", api.R(cmd.Context(), cred).SetQueryParam("up_mid", fmt.Sprintf("%v", info["mid"])), "GET", "https://api.bilibili.com/x/v3/fav/folder/created/list-all")
				if err != nil {
					formatter.ExitError("upstream_error", "获取收藏夹列表失败: "+err.Error(), format)
				}

				var res map[string]interface{}
				json.Unmarshal(data, &res)

				var items []formatter.FavoriteFolder
				if list, ok := res["list"].([]interface{}); ok {
					for _, l := range list {
						lm := l.(map[string]interface{})
						items = append(items, formatter.FavoriteFolder{
							ID:         formatter.ToInt(lm["id"]),
							Title:      fmt.Sprintf("%v", lm["title"]),
							MediaCount: formatter.ToInt(lm["media_count"]),
						})
					}
				}

				if formatter.EmitStructured(formatter.ListPayload{Items: items}, format) {
					return
				}

				fmt.Println("📁 收藏夹列表:")
				for _, i := range items {
					fmt.Printf("- %s  (%d 个内容)  [ID: %d]\n", i.Title, i.MediaCount, i.ID)
				}
				return
			}

			// Get specific folder contents
			folderID := args[0]
			data, err := api.Call(cmd.Context(), "获取收藏夹内容", api.R(cmd.Context(), cred).SetQueryParam("media_id", folderID).SetQueryParam("pn", fmt.Sprintf("%d", page)).SetQueryParam("ps", "20"), "GET", "https://api.bilibili.com/x/v3/fav/resource/list")
			if err != nil {
				formatter.ExitError("upstream_error", "获取收藏夹内容失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			var items []formatter.FavoriteMedia
			if medias, ok := res["medias"].([]interface{}); ok && medias != nil {
				for _, m := range medias {
					mm := m.(map[string]interface{})
					upper := mm["upper"].(map[string]interface{})
					items = append(items, formatter.FavoriteMedia{
						ID:              fmt.Sprintf("%v", mm["bvid"]),
						Bvid:            fmt.Sprintf("%v", mm["bvid"]),
						Title:           fmt.Sprintf("%v", mm["title"]),
						DurationSeconds: formatter.ToInt(mm["duration"]),
						Duration:        formatter.FormatDuration(formatter.ToInt(mm["duration"])),
						Upper: struct {
							Name string `json:"name" yaml:"name"`
						}{
							Name: fmt.Sprintf("%v", upper["name"]),
						},
					})
				}
			}

			payload := map[string]interface{}{
				"medias":   items,
				"has_more": res["has_more"],
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Printf("📁 收藏夹内容 (ID: %s, 第 %d 页):\n\n", folderID, page)
			for i, v := range items {
				fmt.Printf("%d. %s [%s] %s  (UP: %s)\n", i+1, v.Bvid, v.Duration, v.Title, v.Upper.Name)
			}
			if hasMore, ok := res["has_more"].(bool); ok && hasMore {
				fmt.Printf("\n> bili favorites %s --page %d\n", folderID, page+1)
			}
		},
	}
	favoritesCmd.Flags().IntVar(&page, "page", 1, "页码")
	favoritesCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	favoritesCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	followingCmd := &cobra.Command{
		Use:   "following",
		Short: "查看关注列表",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(false, outputJSON, outputYAML)
			api := client.NewClient()

			if page <= 0 {
				formatter.ExitError("invalid_input", "page 必须大于 0", format)
			}

			infoData, err := api.Call(cmd.Context(), "获取当前登录用户信息", api.R(cmd.Context(), cred), "GET", "https://api.bilibili.com/x/space/myinfo")
			if err != nil {
				formatter.ExitError("upstream_error", "获取用户信息失败: "+err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)

			data, err := api.Call(cmd.Context(), "获取关注列表", api.R(cmd.Context(), cred).SetQueryParam("vmid", fmt.Sprintf("%v", info["mid"])).SetQueryParam("pn", fmt.Sprintf("%d", page)).SetQueryParam("ps", "20"), "GET", "https://api.bilibili.com/x/relation/followings")
			if err != nil {
				formatter.ExitError("upstream_error", "获取关注列表失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			var items []formatter.FollowingUser
			if list, ok := res["list"].([]interface{}); ok {
				for _, l := range list {
					lm := l.(map[string]interface{})
					items = append(items, formatter.FollowingUser{
						ID:   fmt.Sprintf("%v", lm["mid"]),
						Name: fmt.Sprintf("%v", lm["uname"]),
						Sign: fmt.Sprintf("%v", lm["sign"]),
					})
				}
			}

			payload := map[string]interface{}{
				"items": items,
				"total": res["total"],
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Printf("👥 关注列表 (第 %d 页):\n\n", page)
			for _, u := range items {
				fmt.Printf("- %s (UID: %s)\n", u.Name, u.ID)
			}
			total := formatter.ToInt(res["total"])
			if page*20 < total {
				fmt.Printf("\n> bili following --page %d\n", page+1)
			}
		},
	}
	followingCmd.Flags().IntVar(&page, "page", 1, "页码")
	followingCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	followingCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "查看观看历史",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(false, outputJSON, outputYAML)
			api := client.NewClient()

			if page <= 0 || maxCount <= 0 || maxCount > 100 {
				formatter.ExitError("invalid_input", "参数非法", format)
			}

			data, err := api.Call(cmd.Context(), "获取观看历史", api.R(cmd.Context(), cred).SetQueryParam("max", "0").SetQueryParam("business", "archive").SetQueryParam("view_at", "0").SetQueryParam("type", "archive").SetQueryParam("pn", fmt.Sprintf("%d", page)).SetQueryParam("ps", fmt.Sprintf("%d", maxCount)), "GET", "https://api.bilibili.com/x/v2/history/toview")
			if err != nil {
				formatter.ExitError("upstream_error", "获取观看历史失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			var items []formatter.HistoryItem
			if list, ok := res["list"].([]interface{}); ok {
				for _, l := range list {
					if len(items) >= maxCount {
						break
					}
					lm := l.(map[string]interface{})
					history := lm["history"].(map[string]interface{})

					author := fmt.Sprintf("%v", lm["author_name"])

					items = append(items, formatter.HistoryItem{
						ID:     fmt.Sprintf("%v", history["bvid"]),
						Bvid:   fmt.Sprintf("%v", history["bvid"]),
						Title:  fmt.Sprintf("%v", lm["title"]),
						Author: author,
					})
				}
			}

			if formatter.EmitStructured(formatter.ListPayload{Items: items}, format) {
				return
			}

			fmt.Printf("🕒 观看历史 (第 %d 页):\n\n", page)
			for i, v := range items {
				fmt.Printf("%d. %s  %s  (UP: %s)\n", i+1, v.Bvid, v.Title, v.Author)
			}
		},
	}
	historyCmd.Flags().IntVar(&page, "page", 1, "页码")
	historyCmd.Flags().IntVar(&maxCount, "max", 30, "每页数量")
	historyCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	historyCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	watchLaterCmd := &cobra.Command{
		Use:   "watch-later",
		Short: "查看稍后再看",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(false, outputJSON, outputYAML)
			api := client.NewClient()

			data, err := api.Call(cmd.Context(), "获取稍后再看", api.R(cmd.Context(), cred), "GET", "https://api.bilibili.com/x/v2/history/toview")
			if err != nil {
				formatter.ExitError("upstream_error", "获取稍后再看失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			var items []formatter.WatchLaterItem
			if list, ok := res["list"].([]interface{}); ok {
				for _, l := range list {
					lm := l.(map[string]interface{})
					owner := lm["owner"].(map[string]interface{})
					items = append(items, formatter.WatchLaterItem{
						ID:              fmt.Sprintf("%v", lm["bvid"]),
						Bvid:            fmt.Sprintf("%v", lm["bvid"]),
						Title:           fmt.Sprintf("%v", lm["title"]),
						DurationSeconds: formatter.ToInt(lm["duration"]),
						Duration:        formatter.FormatDuration(formatter.ToInt(lm["duration"])),
						Author:          fmt.Sprintf("%v", owner["name"]),
					})
				}
			}

			payload := map[string]interface{}{
				"list":  items,
				"count": len(items),
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Printf("🕒 稍后再看 (%d 条):\n\n", len(items))
			for i, v := range items {
				fmt.Printf("%d. %s [%s] %s  (UP: %s)\n", i+1, v.Bvid, v.Duration, v.Title, v.Author)
			}
		},
	}
	watchLaterCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	watchLaterCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rootCmd.AddCommand(favoritesCmd)
	rootCmd.AddCommand(followingCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(watchLaterCmd)
}
