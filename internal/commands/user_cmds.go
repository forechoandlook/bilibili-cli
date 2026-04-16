package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/client"
	"github.com/forechoandlook/bilibili-cli/internal/formatter"
)

func AddUserCommands(rootCmd *cobra.Command) {
	var maxCount int
	var searchType string
	var searchPage int

	userCmd := &cobra.Command{
		Use:   "user <UID或用户名>",
		Short: "查看用户资料",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			api := client.NewClient()
			cred := getOptionalLogin()

			input := args[0]
			uid, err := strconv.ParseInt(input, 10, 64)

			// If not a number, try to search for the user
			if err != nil {
				searchData, err := api.Call(cmd.Context(), "搜索用户", api.R(cmd.Context(), cred).SetQueryParam("keyword", input).SetQueryParam("search_type", "bili_user"), "GET", "https://api.bilibili.com/x/web-interface/search/type")
				if err != nil {
					formatter.ExitError("upstream_error", err.Error(), format)
				}
				var res map[string]interface{}
				json.Unmarshal(searchData, &res)

				if results, ok := res["result"].([]interface{}); ok && len(results) > 0 {
					firstMatch := results[0].(map[string]interface{})
					uid = int64(formatter.ToInt(firstMatch["mid"]))
				} else {
					formatter.ExitError("not_found", fmt.Sprintf("未找到用户: %s", input), format)
				}
			}

			// Get user info
			infoData, err := api.Call(cmd.Context(), "获取用户信息", api.R(cmd.Context(), cred).SetQueryParam("mid", fmt.Sprintf("%d", uid)), "GET", "https://api.bilibili.com/x/space/wbi/acc/info")
			if err != nil {
				formatter.ExitError("upstream_error", err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)

			// Get relation stat
			relData, err := api.Call(cmd.Context(), "获取用户关系信息", api.R(cmd.Context(), cred).SetQueryParam("vmid", fmt.Sprintf("%d", uid)), "GET", "https://api.bilibili.com/x/relation/stat")
			if err != nil {
				formatter.ExitError("upstream_error", err.Error(), format)
			}
			var rel map[string]interface{}
			json.Unmarshal(relData, &rel)

			res := formatter.WhoamiData{
				User: &formatter.UserSummary{
					ID:       fmt.Sprintf("%v", info["mid"]),
					Name:     fmt.Sprintf("%v", info["name"]),
					Username: fmt.Sprintf("%v", info["name"]),
					Level:    formatter.ToInt(info["level"]),
					Coins:    formatter.ToInt(info["coins"]),
					Sign:     fmt.Sprintf("%v", info["sign"]),
				},
				Relation: &formatter.RelationData{
					Following: formatter.ToInt(rel["following"]),
					Follower:  formatter.ToInt(rel["follower"]),
				},
			}

			if formatter.EmitStructured(res, format) {
				return
			}

			fmt.Printf("📺 UP主: %s (UID: %s)\n", res.User.Name, res.User.ID)
			fmt.Printf("等 级: LV%d\n", res.User.Level)
			fmt.Printf("关 注: %d\n", res.Relation.Following)
			fmt.Printf("粉 丝: %s\n", formatter.FormatCount(res.Relation.Follower))
			if res.User.Sign != "" {
				fmt.Printf("签 名: %s\n", res.User.Sign)
			}
		},
	}
	userCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	userCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	userVideosCmd := &cobra.Command{
		Use:   "user-videos <UID>",
		Short: "查看用户视频列表",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			api := client.NewClient()
			cred := getOptionalLogin()

			uid := args[0]

			if maxCount <= 0 {
				formatter.ExitError("invalid_input", "max 必须大于 0", format)
			}

			// Simply fetch first page up to maxCount for now.
			perPage := maxCount
			if perPage > 50 {
				perPage = 50
			}

			data, err := api.Call(cmd.Context(), "获取视频列表", api.R(cmd.Context(), cred).SetQueryParam("mid", uid).SetQueryParam("pn", "1").SetQueryParam("ps", fmt.Sprintf("%d", perPage)), "GET", "https://api.bilibili.com/x/space/wbi/arc/search")
			if err != nil {
				formatter.ExitError("upstream_error", "获取视频列表失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			list := res["list"].(map[string]interface{})
			vlist := list["vlist"].([]interface{})

			var results []formatter.VideoSummary
			for _, v := range vlist {
				if len(results) >= maxCount {
					break
				}
				vi := v.(map[string]interface{})
				results = append(results, formatter.VideoSummary{
					ID:       fmt.Sprintf("%v", vi["bvid"]),
					Bvid:     fmt.Sprintf("%v", vi["bvid"]),
					Title:    fmt.Sprintf("%v", vi["title"]),
					Duration: fmt.Sprintf("%v", vi["length"]), // format from api is like "01:23"
					Stats: formatter.VideoStats{
						View: formatter.ToInt(vi["play"]),
					},
				})
			}

			payload := formatter.ListPayload{Items: results}
			if formatter.EmitStructured(payload, format) {
				return
			}

			for i, v := range results {
				fmt.Printf("%d. %s [%s] %s\n", i+1, v.Bvid, formatter.FormatCount(v.Stats.View), v.Title)
			}
		},
	}
	userVideosCmd.Flags().IntVar(&maxCount, "max", 10, "最大返回数量")
	userVideosCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	userVideosCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	searchCmd := &cobra.Command{
		Use:   "search <关键词>",
		Short: "搜索用户或视频",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			api := client.NewClient()
			cred := getOptionalLogin()

			keyword := args[0]

			if maxCount <= 0 {
				formatter.ExitError("invalid_input", "max 必须大于 0", format)
			}
			if searchPage <= 0 {
				formatter.ExitError("invalid_input", "page 必须大于 0", format)
			}

			var searchTypeAPI string
			if searchType == "user" {
				searchTypeAPI = "bili_user"
			} else if searchType == "video" {
				searchTypeAPI = "video"
			} else {
				formatter.ExitError("invalid_input", "type 必须是 user 或 video", format)
			}

			data, err := api.Call(cmd.Context(), "搜索", api.R(cmd.Context(), cred).SetQueryParam("keyword", keyword).SetQueryParam("search_type", searchTypeAPI).SetQueryParam("page", fmt.Sprintf("%d", searchPage)), "GET", "https://api.bilibili.com/x/web-interface/search/type")
			if err != nil {
				if searchType == "user" {
					formatter.ExitError("upstream_error", "搜索用户失败: "+err.Error(), format)
				} else {
					formatter.ExitError("upstream_error", "搜索视频失败: "+err.Error(), format)
				}
			}

			var res map[string]interface{}
			json.Unmarshal(data, &res)

			results := res["result"]
			if results == nil {
				if format != formatter.OutputFormatNone {
					formatter.EmitStructured(formatter.ListPayload{Items: []interface{}{}}, format)
					return
				}
				fmt.Println("未找到结果。")
				return
			}

			resultList := results.([]interface{})

			if searchType == "user" {
				var items []formatter.SearchUser
				for _, r := range resultList {
					if len(items) >= maxCount {
						break
					}
					rm := r.(map[string]interface{})
					items = append(items, formatter.SearchUser{
						ID:     fmt.Sprintf("%v", rm["mid"]),
						Name:   fmt.Sprintf("%v", rm["uname"]),
						Sign:   fmt.Sprintf("%v", rm["usign"]),
						Fans:   formatter.ToInt(rm["fans"]),
						Videos: formatter.ToInt(rm["videos"]),
					})
				}
				if formatter.EmitStructured(formatter.ListPayload{Items: items}, format) {
					return
				}
				if len(items) == 0 {
					fmt.Println("未找到结果。")
					return
				}
				for _, u := range items {
					fmt.Printf("- %s (UID: %s)  粉: %s  视频: %d\n", u.Name, u.ID, formatter.FormatCount(u.Fans), u.Videos)
				}

			} else {
				var items []formatter.SearchVideo
				for _, r := range resultList {
					if len(items) >= maxCount {
						break
					}
					rm := r.(map[string]interface{})
					bvid := fmt.Sprintf("%v", rm["bvid"])
					if bvid == "" {
						continue // API sometimes returns empty bvid for pseudo-videos
					}
					items = append(items, formatter.SearchVideo{
						ID:       bvid,
						Bvid:     bvid,
						Title:    fmt.Sprintf("%v", rm["title"]), // Note: raw title might have HTML highlights
						Author:   fmt.Sprintf("%v", rm["author"]),
						Play:     formatter.ToInt(rm["play"]),
						Duration: fmt.Sprintf("%v", rm["duration"]),
					})
				}
				if formatter.EmitStructured(formatter.ListPayload{Items: items}, format) {
					return
				}
				if len(items) == 0 {
					fmt.Println("未找到结果。")
					return
				}
				for i, v := range items {
					fmt.Printf("%d. %s [%s] %s  (UP: %s)\n", i+1, v.Bvid, formatter.FormatCount(v.Play), v.Title, v.Author)
				}
			}
		},
	}
	searchCmd.Flags().StringVar(&searchType, "type", "user", "搜索类型：user 或 video")
	searchCmd.Flags().IntVar(&maxCount, "max", 10, "最大返回数量")
	searchCmd.Flags().IntVar(&searchPage, "page", 1, "页码")
	searchCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	searchCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(userVideosCmd)
	rootCmd.AddCommand(searchCmd)
}
