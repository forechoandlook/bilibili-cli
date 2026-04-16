package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/client"
	"github.com/forechoandlook/bilibili-cli/internal/formatter"
)

func AddDiscoveryCommands(rootCmd *cobra.Command) {
	var maxCount int
	var page int
	var day int
	var offset string

	hotCmd := &cobra.Command{
		Use:   "hot",
		Short: "查看热门视频",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			api := client.NewClient()

			if maxCount <= 0 {
				formatter.ExitError("invalid_input", "max 必须大于 0", format)
			}
			if page <= 0 {
				formatter.ExitError("invalid_input", "page 必须大于 0", format)
			}

			// For hot, ps is page size and pn is page number
			data, err := api.Call(cmd.Context(), "获取热门视频", api.R(cmd.Context(), nil).SetQueryParam("pn", fmt.Sprintf("%d", page)).SetQueryParam("ps", fmt.Sprintf("%d", maxCount)), "GET", "https://api.bilibili.com/x/web-interface/popular")
			if err != nil {
				formatter.ExitError("upstream_error", "获取热门视频失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			list := res["list"].([]interface{})
			var items []formatter.VideoSummary
			for _, v := range list {
				if len(items) >= maxCount {
					break
				}
				vi := v.(map[string]interface{})
				owner := vi["owner"].(map[string]interface{})
				stat := vi["stat"].(map[string]interface{})

				items = append(items, formatter.VideoSummary{
					ID:       fmt.Sprintf("%v", vi["bvid"]),
					Bvid:     fmt.Sprintf("%v", vi["bvid"]),
					Title:    fmt.Sprintf("%v", vi["title"]),
					Owner: formatter.VideoOwner{
						Name: fmt.Sprintf("%v", owner["name"]),
					},
					Stats: formatter.VideoStats{
						View: formatter.ToInt(stat["view"]),
						Like: formatter.ToInt(stat["like"]),
					},
				})
			}

			payload := formatter.ListPayload{Items: items}
			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Printf("🔥 热门视频 (第 %d 页):\n\n", page)
			for i, v := range items {
				fmt.Printf("%d. %s [%s] %s  (UP: %s)\n", i+1, v.Bvid, formatter.FormatCount(v.Stats.View), v.Title, v.Owner.Name)
			}
		},
	}
	hotCmd.Flags().IntVar(&maxCount, "max", 10, "最大返回数量")
	hotCmd.Flags().IntVar(&page, "page", 1, "页码")
	hotCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	hotCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rankCmd := &cobra.Command{
		Use:   "rank",
		Short: "查看全站排行榜",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			api := client.NewClient()

			if maxCount <= 0 {
				formatter.ExitError("invalid_input", "max 必须大于 0", format)
			}

			dayType := 3
			if day == 7 {
				dayType = 7
			}

			data, err := api.Call(cmd.Context(), "获取排行榜", api.R(cmd.Context(), nil).SetQueryParam("day", fmt.Sprintf("%d", dayType)), "GET", "https://api.bilibili.com/x/web-interface/ranking/v2")
			if err != nil {
				formatter.ExitError("upstream_error", "获取排行榜失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			list := res["list"].([]interface{})
			var items []formatter.VideoSummary
			for _, v := range list {
				if len(items) >= maxCount {
					break
				}
				vi := v.(map[string]interface{})
				owner := vi["owner"].(map[string]interface{})
				stat := vi["stat"].(map[string]interface{})

				items = append(items, formatter.VideoSummary{
					ID:       fmt.Sprintf("%v", vi["bvid"]),
					Bvid:     fmt.Sprintf("%v", vi["bvid"]),
					Title:    fmt.Sprintf("%v", vi["title"]),
					Owner: formatter.VideoOwner{
						Name: fmt.Sprintf("%v", owner["name"]),
					},
					Stats: formatter.VideoStats{
						View: formatter.ToInt(stat["view"]),
					},
				})
			}

			payload := formatter.ListPayload{Items: items}
			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Printf("🏆 全站排行榜 (%d日榜):\n\n", dayType)
			for i, v := range items {
				fmt.Printf("%d. %s [%s] %s  (UP: %s)\n", i+1, v.Bvid, formatter.FormatCount(v.Stats.View), v.Title, v.Owner.Name)
			}
		},
	}
	rankCmd.Flags().IntVar(&maxCount, "max", 10, "最大返回数量")
	rankCmd.Flags().IntVar(&day, "day", 3, "榜单天数: 3 或 7")
	rankCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	rankCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	feedCmd := &cobra.Command{
		Use:   "feed",
		Short: "查看动态时间线",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(false, outputJSON, outputYAML)
			api := client.NewClient()

			data, err := api.Call(cmd.Context(), "获取动态", api.R(cmd.Context(), cred).SetQueryParam("offset", offset).SetQueryParam("update_baseline", "0"), "GET", "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all")
			if err != nil {
				formatter.ExitError("upstream_error", "获取动态失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			var items []formatter.DynamicItem
			if itemList, ok := res["items"].([]interface{}); ok {
				for _, i := range itemList {
					im := i.(map[string]interface{})
					mod := im["modules"].(map[string]interface{})
					modAuth := mod["module_author"].(map[string]interface{})
					modDyn := mod["module_dynamic"].(map[string]interface{})

					var text string
					var title string

					if desc, ok := modDyn["desc"].(map[string]interface{}); ok {
						text = fmt.Sprintf("%v", desc["text"])
					}

					if major, ok := modDyn["major"].(map[string]interface{}); ok && major != nil {
						if archive, ok := major["archive"].(map[string]interface{}); ok {
							title = fmt.Sprintf("%v", archive["title"])
						} else if article, ok := major["article"].(map[string]interface{}); ok {
							title = fmt.Sprintf("%v", article["title"])
						}
					}

					idStr := fmt.Sprintf("%v", im["id_str"])

					items = append(items, formatter.DynamicItem{
						ID: idStr,
						Author: struct {
							Name string `json:"name" yaml:"name"`
						}{
							Name: fmt.Sprintf("%v", modAuth["name"]),
						},
						PublishedLabel: fmt.Sprintf("%v", modAuth["pub_time"]),
						Title: title,
						Text: text,
					})
				}
			}

			payload := map[string]interface{}{
				"items":       items,
				"next_offset": res["offset"],
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Println("📰 动态时间线:")
			for _, d := range items {
				fmt.Printf("[cyan]%s[/cyan]  [dim]%s[/dim]  (ID: %s)\n", d.Author.Name, d.PublishedLabel, d.ID)
				if d.Title != "" {
					fmt.Printf("标题: %s\n", d.Title)
				}
				if d.Text != "" {
					text := d.Text
					if len(text) > 100 {
						text = string([]rune(text)[:100]) + "..."
					}
					fmt.Printf("内容: %s\n", text)
				}
				fmt.Println()
			}

			if nextOffset := fmt.Sprintf("%v", res["offset"]); nextOffset != "" && nextOffset != "0" {
				fmt.Printf("--- 下一页 ---\n> bili feed --offset %s\n", nextOffset)
			}
		},
	}
	feedCmd.Flags().StringVar(&offset, "offset", "", "翻页游标")
	feedCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	feedCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rootCmd.AddCommand(hotCmd)
	rootCmd.AddCommand(rankCmd)
	rootCmd.AddCommand(feedCmd)
}
