package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/client"
	"github.com/forechoandlook/bilibili-cli/internal/formatter"
)

func AddVideoCommands(rootCmd *cobra.Command) {
	var (
		subtitle         bool
		subtitleTimeline bool
		subtitleFormat   string
		comments         bool
		ai               bool
		related          bool
	)

	videoCmd := &cobra.Command{
		Use:   "video <BV号或URL>",
		Short: "查看视频详情",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)

			bvid, err := client.ExtractBvid(args[0])
			if err != nil {
				formatter.ExitError("invalid_input", err.Error(), format)
			}

			api := client.NewClient()
			cred := getOptionalLogin()

			// 1. Get Video Info
			infoData, err := api.Call(cmd.Context(), "获取视频信息", api.R(cmd.Context(), cred).SetQueryParam("bvid", bvid), "GET", "https://api.bilibili.com/x/web-interface/view")
			if err != nil {
				formatter.ExitError("upstream_error", err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)

			owner := info["owner"].(map[string]interface{})
			stat := info["stat"].(map[string]interface{})

			videoSummary := &formatter.VideoSummary{
				ID:              bvid,
				Bvid:            bvid,
				Aid:             formatter.ToInt(info["aid"]),
				Title:           fmt.Sprintf("%v", info["title"]),
				Description:     fmt.Sprintf("%v", info["desc"]),
				DurationSeconds: formatter.ToInt(info["duration"]),
				Duration:        formatter.FormatDuration(formatter.ToInt(info["duration"])),
				URL:             fmt.Sprintf("https://www.bilibili.com/video/%s", bvid),
				Owner: formatter.VideoOwner{
					ID:   fmt.Sprintf("%v", owner["mid"]),
					Name: fmt.Sprintf("%v", owner["name"]),
				},
				Stats: formatter.VideoStats{
					View:     formatter.ToInt(stat["view"]),
					Danmaku:  formatter.ToInt(stat["danmaku"]),
					Like:     formatter.ToInt(stat["like"]),
					Coin:     formatter.ToInt(stat["coin"]),
					Favorite: formatter.ToInt(stat["favorite"]),
					Share:    formatter.ToInt(stat["share"]),
				},
			}

			payload := formatter.VideoCommandPayload{
				Video:    videoSummary,
				Warnings: []formatter.Warning{},
			}

			// Subtitle logic (simplified for Go version)
			if subtitle || subtitleTimeline {
				payload.Warnings = append(payload.Warnings, formatter.Warning{Code: "subtitle_unavailable", Message: "获取字幕在当前 Go 版本中未实现"})
			}

			// AI logic
			if ai {
				payload.Warnings = append(payload.Warnings, formatter.Warning{Code: "ai_summary_unavailable", Message: "获取 AI 总结在当前 Go 版本中未实现"})
			}

			// Comments logic
			if comments {
				cmData, err := api.Call(cmd.Context(), "获取评论", api.R(cmd.Context(), cred).SetQueryParam("oid", fmt.Sprintf("%d", videoSummary.Aid)).SetQueryParam("type", "1").SetQueryParam("sort", "2"), "GET", "https://api.bilibili.com/x/v2/reply")
				if err != nil {
					payload.Warnings = append(payload.Warnings, formatter.Warning{Code: "comments_unavailable", Message: "获取评论失败"})
				} else {
					var cmResp map[string]interface{}
					if err := json.Unmarshal(cmData, &cmResp); err == nil {
						if replies, ok := cmResp["replies"].([]interface{}); ok {
							for _, r := range replies {
								rm := r.(map[string]interface{})
								member := rm["member"].(map[string]interface{})
								content := rm["content"].(map[string]interface{})

								payload.Comments = append(payload.Comments, formatter.Comment{
									ID: fmt.Sprintf("%v", rm["rpid"]),
									Author: formatter.VideoOwner{
										ID:   fmt.Sprintf("%v", member["mid"]),
										Name: fmt.Sprintf("%v", member["uname"]),
									},
									Message:    fmt.Sprintf("%v", content["message"]),
									Like:       formatter.ToInt(rm["like"]),
									ReplyCount: formatter.ToInt(rm["rcount"]),
								})
							}
						}
					}
				}
			}

			// Related logic
			if related {
				relData, err := api.Call(cmd.Context(), "获取相关推荐", api.R(cmd.Context(), cred).SetQueryParam("bvid", bvid), "GET", "https://api.bilibili.com/x/web-interface/archive/related")
				if err != nil {
					payload.Warnings = append(payload.Warnings, formatter.Warning{Code: "related_unavailable", Message: "获取相关推荐失败"})
				} else {
					var relList []interface{}
					if err := json.Unmarshal(relData, &relList); err == nil {
						for _, item := range relList {
							ri := item.(map[string]interface{})
							ro := ri["owner"].(map[string]interface{})
							rs := ri["stat"].(map[string]interface{})
							payload.Related = append(payload.Related, formatter.VideoSummary{
								ID:              fmt.Sprintf("%v", ri["bvid"]),
								Bvid:            fmt.Sprintf("%v", ri["bvid"]),
								Aid:             formatter.ToInt(ri["aid"]),
								Title:           fmt.Sprintf("%v", ri["title"]),
								DurationSeconds: formatter.ToInt(ri["duration"]),
								Duration:        formatter.FormatDuration(formatter.ToInt(ri["duration"])),
								Owner: formatter.VideoOwner{
									ID:   fmt.Sprintf("%v", ro["mid"]),
									Name: fmt.Sprintf("%v", ro["name"]),
								},
								Stats: formatter.VideoStats{
									View: formatter.ToInt(rs["view"]),
								},
							})
						}
					}
				}
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			// Console output
			fmt.Printf("📺 %s\n", videoSummary.Title)
			fmt.Printf("BV号: %s\n", videoSummary.Bvid)
			fmt.Printf("UP主: %s  (UID: %s)\n", videoSummary.Owner.Name, videoSummary.Owner.ID)
			fmt.Printf("时长: %s\n", videoSummary.Duration)
			fmt.Printf("播放: %s\n", formatter.FormatCount(videoSummary.Stats.View))
			fmt.Printf("弹幕: %s\n", formatter.FormatCount(videoSummary.Stats.Danmaku))
			fmt.Printf("点赞: %s\n", formatter.FormatCount(videoSummary.Stats.Like))
			fmt.Printf("投币: %s\n", formatter.FormatCount(videoSummary.Stats.Coin))
			fmt.Printf("收藏: %s\n", formatter.FormatCount(videoSummary.Stats.Favorite))
			fmt.Printf("链接: %s\n", videoSummary.URL)

			if videoSummary.Description != "" {
				desc := videoSummary.Description
				if len(desc) > 200 {
					desc = string([]rune(desc)[:200]) + "..."
				}
				fmt.Printf("简介: %s\n", desc)
			}

			if len(payload.Comments) > 0 {
				fmt.Println("\n💬 热门评论:")
				for i, c := range payload.Comments {
					if i >= 10 {
						break
					}
					fmt.Printf("  %s (👍 %d)\n", c.Author.Name, c.Like)

					msg := c.Message
					if len(msg) > 120 {
						msg = string([]rune(msg)[:120]) + "..."
					}
					fmt.Printf("  %s\n\n", msg)
				}
			}

			if len(payload.Related) > 0 {
				fmt.Println("\n📎 相关推荐:")
				for i, r := range payload.Related {
					if i >= 10 {
						break
					}
					fmt.Printf("%d. %s [%s] %s\n", i+1, r.Bvid, formatter.FormatCount(r.Stats.View), r.Title)
				}
			}
		},
	}

	videoCmd.Flags().BoolVarP(&subtitle, "subtitle", "s", false, "显示字幕内容。")
	videoCmd.Flags().BoolVarP(&subtitleTimeline, "subtitle-timeline", "", false, "显示带时间线的字幕。")
	videoCmd.Flags().StringVarP(&subtitleFormat, "subtitle-format", "", "timeline", "字幕格式：timeline 或 srt。")
	videoCmd.Flags().BoolVarP(&comments, "comments", "c", false, "显示评论。")
	videoCmd.Flags().BoolVarP(&ai, "ai", "", false, "显示 AI 总结。")
	videoCmd.Flags().BoolVarP(&related, "related", "r", false, "显示相关推荐视频。")
	videoCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	videoCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rootCmd.AddCommand(videoCmd)
}
