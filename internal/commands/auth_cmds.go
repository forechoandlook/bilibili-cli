package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/auth"
	"github.com/forechoandlook/bilibili-cli/internal/client"
	"github.com/forechoandlook/bilibili-cli/internal/formatter"
)

var (
	outputJSON bool
	outputYAML bool
)

func requireLogin(requireWrite bool, asJSON, asYAML bool) *auth.Credential {
	format := formatter.ResolveOutputFormat(asJSON, asYAML)
	mode := auth.AuthModeRead
	if requireWrite {
		mode = auth.AuthModeWrite
	}

	cred, err := auth.GetCredential(mode)
	if err != nil {
		if format != formatter.OutputFormatNone {
			formatter.EmitError("not_authenticated", "未登录。使用 bili login 登录。", nil, format)
		} else {
			fmt.Println("⚠️  需要登录。使用 bili login 登录。")
		}
		panic("exit:1") // Will be caught by main wrapper
	}
	return cred
}

func getOptionalLogin() *auth.Credential {
	cred, _ := auth.GetCredential(auth.AuthModeOptional)
	return cred
}

func AddAuthCommands(rootCmd *cobra.Command) {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "检查登录状态",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred, err := auth.GetCredential(auth.AuthModeRead)

			if err != nil || cred == nil {
				if format != formatter.OutputFormatNone {
					formatter.EmitError("not_authenticated", "未登录。使用 bili login 登录。", nil, format)
				} else {
					fmt.Println("❌ 未登录。")
				}
				panic("exit:1")
			}

			// Try to get self info
			api := client.NewClient()
			data, err := api.Call(cmd.Context(), "获取当前登录用户信息", api.R(cmd.Context(), cred), "GET", "https://api.bilibili.com/x/space/myinfo")
			if err != nil {
				if format != formatter.OutputFormatNone {
					formatter.EmitError("network_error", "获取用户信息失败", nil, format)
				} else {
					fmt.Println("❌ 获取用户信息失败")
				}
				panic("exit:1")
			}

			var info map[string]interface{}
			json.Unmarshal(data, &info)

			userSummary := &formatter.UserSummary{
				ID:       fmt.Sprintf("%v", info["mid"]),
				Name:     fmt.Sprintf("%v", info["name"]),
				Username: fmt.Sprintf("%v", info["name"]),
				Level:    formatter.ToInt(info["level"]),
				Coins:    formatter.ToInt(info["coins"]),
			}

			res := formatter.AuthStatusData{
				Authenticated: true,
				User:          userSummary,
			}

			if formatter.EmitStructured(res, format) {
				return
			}

			fmt.Printf("✅ 已登录\n")
			fmt.Printf("UP主: %s (UID: %s)\n", userSummary.Name, userSummary.ID)
		},
	}
	statusCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	statusCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "扫码登录",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := auth.QRLogin()
			if err != nil {
				fmt.Printf("❌ 登录失败: %v\n", err)
				panic("exit:1")
			}
		},
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "注销登录",
		Run: func(cmd *cobra.Command, args []string) {
			auth.ClearCredential()
			fmt.Println("✅ 已注销")
		},
	}

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "查看个人信息",
		Run: func(cmd *cobra.Command, args []string) {
			format := formatter.ResolveOutputFormat(outputJSON, outputYAML)
			cred := requireLogin(false, outputJSON, outputYAML)

			api := client.NewClient()

			// Get Self Info
			infoData, err := api.Call(cmd.Context(), "获取用户信息", api.R(cmd.Context(), cred), "GET", "https://api.bilibili.com/x/space/myinfo")
			if err != nil {
				formatter.ExitError("upstream_error", err.Error(), format)
			}
			var info map[string]interface{}
			json.Unmarshal(infoData, &info)

			// Get Relation Info
			relData, err := api.Call(cmd.Context(), "获取关系信息", api.R(cmd.Context(), cred).SetQueryParam("vmid", fmt.Sprintf("%v", info["mid"])), "GET", "https://api.bilibili.com/x/relation/stat")
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
			fmt.Printf("硬 币: %d\n", res.User.Coins)
			fmt.Printf("关 注: %d\n", res.Relation.Following)
			fmt.Printf("粉 丝: %s\n", formatter.FormatCount(res.Relation.Follower))
			if res.User.Sign != "" {
				fmt.Printf("签 名: %s\n", res.User.Sign)
			}
		},
	}
	whoamiCmd.Flags().BoolVar(&outputJSON, "json", false, "输出 JSON")
	whoamiCmd.Flags().BoolVar(&outputYAML, "yaml", false, "输出 YAML")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}
