package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/client"
	"github.com/forechoandlook/bilibili-cli/internal/formatter"
)

func AddInteractionCommands(rootCmd *cobra.Command) {
	var num int
	var yes bool

	likeCmd := &cobra.Command{
		Use:   "like <BV号>",
		Short: "点赞视频",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(true, outputJSON, outputYAML)
			api := client.NewClient()

			bvid, err := client.ExtractBvid(args[0])
			if err != nil {
				formatter.ExitError("invalid_input", err.Error(), format)
			}

			// Get aid
			infoData, err := api.Call(cmd.Context(), "获取视频信息", api.R(cmd.Context(), cred).SetQueryParam("bvid", bvid), "GET", "https://api.bilibili.com/x/web-interface/view")
			if err != nil {
				formatter.ExitError("upstream_error", "获取视频信息失败: "+err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)
			aid := fmt.Sprintf("%v", info["aid"])

			// Like
			_, err = api.Call(cmd.Context(), "点赞", api.R(cmd.Context(), cred).SetFormData(map[string]string{
				"aid":  aid,
				"like": "1",
				"csrf": cred.BiliJCT,
			}), "POST", "https://api.bilibili.com/x/web-interface/archive/like")

			if err != nil {
				formatter.ExitError("upstream_error", "操作失败: "+err.Error(), format)
			}

			payload := formatter.ActionResult{
				Success: true,
				Action:  "like",
				Result: map[string]interface{}{
					"bvid": bvid,
				},
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Println("✅ 点赞成功")
		},
	}
	likeCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	likeCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	coinCmd := &cobra.Command{
		Use:   "coin <BV号>",
		Short: "投币",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(true, outputJSON, outputYAML)
			api := client.NewClient()

			bvid, err := client.ExtractBvid(args[0])
			if err != nil {
				formatter.ExitError("invalid_input", err.Error(), format)
			}

			if num != 1 && num != 2 {
				formatter.ExitError("invalid_input", "投币数量必须是 1 或 2", format)
			}

			// Get aid
			infoData, err := api.Call(cmd.Context(), "获取视频信息", api.R(cmd.Context(), cred).SetQueryParam("bvid", bvid), "GET", "https://api.bilibili.com/x/web-interface/view")
			if err != nil {
				formatter.ExitError("upstream_error", "获取视频信息失败: "+err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)
			aid := fmt.Sprintf("%v", info["aid"])

			// Coin
			_, err = api.Call(cmd.Context(), "投币", api.R(cmd.Context(), cred).SetFormData(map[string]string{
				"aid":      aid,
				"multiply": fmt.Sprintf("%d", num),
				"select_like": "0",
				"csrf":     cred.BiliJCT,
			}), "POST", "https://api.bilibili.com/x/web-interface/coin/add")

			if err != nil {
				formatter.ExitError("upstream_error", "投币失败: "+err.Error(), format)
			}

			payload := formatter.ActionResult{
				Success: true,
				Action:  "coin",
				Result: map[string]interface{}{
					"bvid":  bvid,
					"coins": num,
				},
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Println("✅ 投币成功")
		},
	}
	coinCmd.Flags().IntVar(&num, "num", 1, "投币数量 (1 或 2)")
	coinCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	coinCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	tripleCmd := &cobra.Command{
		Use:   "triple <BV号>",
		Short: "一键三连",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(true, outputJSON, outputYAML)
			api := client.NewClient()

			bvid, err := client.ExtractBvid(args[0])
			if err != nil {
				formatter.ExitError("invalid_input", err.Error(), format)
			}

			// Get aid
			infoData, err := api.Call(cmd.Context(), "获取视频信息", api.R(cmd.Context(), cred).SetQueryParam("bvid", bvid), "GET", "https://api.bilibili.com/x/web-interface/view")
			if err != nil {
				formatter.ExitError("upstream_error", "获取视频信息失败: "+err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)
			aid := fmt.Sprintf("%v", info["aid"])

			// Triple
			resData, err := api.Call(cmd.Context(), "一键三连", api.R(cmd.Context(), cred).SetFormData(map[string]string{
				"aid":  aid,
				"csrf": cred.BiliJCT,
			}), "POST", "https://api.bilibili.com/x/web-interface/archive/like/triple")

			if err != nil {
				formatter.ExitError("upstream_error", "三连失败: "+err.Error(), format)
			}

			var res map[string]interface{}
			json.Unmarshal(resData, &res)

			payload := formatter.ActionResult{
				Success: true,
				Action:  "triple",
				Result: map[string]interface{}{
					"bvid":     bvid,
					"like":     res["like"],
					"coin":     res["coin"],
					"favorite": res["fav"],
				},
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Println("🎉 一键三连成功")
		},
	}
	tripleCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	tripleCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	unfollowCmd := &cobra.Command{
		Use:   "unfollow <UID>",
		Short: "取消关注",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(true, outputJSON, outputYAML)
			api := client.NewClient()

			uid := args[0]

			if !yes && format == formatter.OutputFormatNone {
				var confirm string
				fmt.Printf("确定要取消关注用户 %s 吗？(y/N): ", uid)
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("已取消")
					return
				}
			}

			_, err := api.Call(cmd.Context(), "取消关注", api.R(cmd.Context(), cred).SetFormData(map[string]string{
				"fid":  uid,
				"act":  "2", // 2 = unsubscribe
				"re_src": "11",
				"csrf": cred.BiliJCT,
			}), "POST", "https://api.bilibili.com/x/relation/modify")

			if err != nil {
				formatter.ExitError("upstream_error", "取消关注失败: "+err.Error(), format)
			}

			payload := formatter.ActionResult{
				Success: true,
				Action:  "unfollow",
				Result: map[string]interface{}{
					"uid": uid,
				},
			}

			if formatter.EmitStructured(payload, format) {
				return
			}

			fmt.Println("✅ 已取消关注")
		},
	}
	unfollowCmd.Flags().BoolVar(&yes, "yes", false, "跳过确认提示")
	unfollowCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	unfollowCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rootCmd.AddCommand(likeCmd)
	rootCmd.AddCommand(coinCmd)
	rootCmd.AddCommand(tripleCmd)
	rootCmd.AddCommand(unfollowCmd)
}
