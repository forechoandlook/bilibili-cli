package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/go-resty/resty/v2"
)

type AuthMode string

const (
	AuthModeOptional AuthMode = "optional"
	AuthModeRead     AuthMode = "read"
	AuthModeWrite    AuthMode = "write"
)

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".bilibili-cli"
	}
	return filepath.Join(home, ".bilibili-cli")
}

func getCredentialFile() string {
	return filepath.Join(getConfigDir(), "credential.json")
}

func LoadCredential() (*Credential, error) {
	data, err := os.ReadFile(getCredentialFile())
	if err != nil {
		return nil, err
	}
	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, err
	}
	if cred.SESSDATA == "" {
		return nil, fmt.Errorf("invalid credential")
	}
	return &cred, nil
}

func SaveCredential(cred *Credential) error {
	dir := getConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cred.SavedAt = time.Now().Unix()
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getCredentialFile(), data, 0600)
}

func ClearCredential() error {
	err := os.Remove(getCredentialFile())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func GetCredential(mode AuthMode) (*Credential, error) {
	cred, err := LoadCredential()
	if err != nil {
		if mode == AuthModeOptional {
			return nil, nil
		}
		return nil, fmt.Errorf("not_authenticated: no credential found")
	}

	if mode == AuthModeWrite && cred.BiliJCT == "" {
		return nil, fmt.Errorf("permission_denied: missing bili_jct for write operation")
	}

	// Simplification: In Go version, skip browser extraction and TTL refresh logic for now
	// to focus on core CLI operations. Assume existing credential is valid unless proven otherwise
	// during API calls.
	return cred, nil
}

// QR code login flow structures
type qrcodeGenerateResp struct {
	Code int `json:"code"`
	Data struct {
		Url      string `json:"url"`
		QrcodeKey string `json:"qrcode_key"`
	} `json:"data"`
}

type qrcodePollResp struct {
	Code int `json:"code"`
	Data struct {
		Url          string `json:"url"`
		RefreshToken string `json:"refresh_token"`
		Timestamp    int64  `json:"timestamp"`
		Code         int    `json:"code"`
		Message      string `json:"message"`
	} `json:"data"`
}

func QRLogin() (*Credential, error) {
	client := resty.New()
	client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")

	// 1. Generate QR Code
	var genResp qrcodeGenerateResp
	_, err := client.R().SetResult(&genResp).Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")
	if err != nil {
		return nil, fmt.Errorf("network_error: %v", err)
	}
	if genResp.Code != 0 {
		return nil, fmt.Errorf("upstream_error: %d", genResp.Code)
	}

	fmt.Println("\n📱 请使用 Bilibili App 扫描以下二维码登录:")
	qrterminal.GenerateHalfBlock(genResp.Data.Url, qrterminal.L, os.Stdout)
	fmt.Println("\n⭐ 扫码后请在手机上确认登录...")

	// 2. Poll status
	pollUrl := fmt.Sprintf("https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=%s", genResp.Data.QrcodeKey)

	for {
		var pollData qrcodePollResp
		rResp, err := client.R().SetResult(&pollData).Get(pollUrl)
		if err != nil {
			return nil, fmt.Errorf("network_error: %v", err)
		}

		switch pollData.Data.Code {
		case 0: // Success
			// Extract cookies
			var sessdata, biliJCT, dedeuserid string
			for _, cookie := range rResp.Cookies() {
				switch cookie.Name {
				case "SESSDATA":
					sessdata = cookie.Value
				case "bili_jct":
					biliJCT = cookie.Value
				case "DedeUserID":
					dedeuserid = cookie.Value
				}
			}
			cred := &Credential{
				SESSDATA:   sessdata,
				BiliJCT:    biliJCT,
				DedeUserID: dedeuserid,
			}
			SaveCredential(cred)
			fmt.Println("\n✅ 登录成功！凭证已保存")
			return cred, nil
		case 86101: // Not scanned
			// wait
		case 86090: // Scanned, wait for confirmation
			fmt.Println("  📲 已扫码，请在手机上确认...")
		case 86038: // Expired
			return nil, fmt.Errorf("二维码已过期，请重试")
		default:
			return nil, fmt.Errorf("未知状态: %d - %s", pollData.Data.Code, pollData.Data.Message)
		}
		time.Sleep(2 * time.Second)
	}
}
