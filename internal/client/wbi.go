package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/forechoandlook/bilibili-cli/internal/auth"
)

var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

func getMixinKey(orig string) string {
	var s strings.Builder
	for _, i := range mixinKeyEncTab {
		if i < len(orig) {
			s.WriteByte(orig[i])
		}
	}
	return s.String()[:32]
}

var (
	cachedImgKey string
	cachedSubKey string
	cacheExpiry  time.Time
)

func (c *APIClient) getWbiKeys(ctx context.Context, cred *auth.Credential) (string, string, error) {
	if time.Now().Before(cacheExpiry) && cachedImgKey != "" && cachedSubKey != "" {
		return cachedImgKey, cachedSubKey, nil
	}

	req := c.R(ctx, cred)
	resp, err := req.Execute("GET", "https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return "", "", err
	}

	var data struct {
		Code int `json:"code"`
		Data struct {
			WbiImg struct {
				ImgUrl string `json:"img_url"`
				SubUrl string `json:"sub_url"`
			} `json:"wbi_img"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return "", "", err
	}
	if data.Code != 0 {
		// Even if not logged in (-101), nav API still returns wbi_img keys in data
		if data.Data.WbiImg.ImgUrl == "" || data.Data.WbiImg.SubUrl == "" {
			return "", "", fmt.Errorf("nav API returned %d and no wbi keys", data.Code)
		}
	}

	extractKey := func(u string) string {
		parts := strings.Split(u, "/")
		if len(parts) == 0 {
			return ""
		}
		filename := parts[len(parts)-1]
		return strings.Split(filename, ".")[0]
	}

	cachedImgKey = extractKey(data.Data.WbiImg.ImgUrl)
	cachedSubKey = extractKey(data.Data.WbiImg.SubUrl)
	cacheExpiry = time.Now().Add(time.Hour)

	return cachedImgKey, cachedSubKey, nil
}

// SignWbi generates WBI signature for parameters
func (c *APIClient) SignWbi(ctx context.Context, cred *auth.Credential, params map[string]string) (string, error) {
	imgKey, subKey, err := c.getWbiKeys(ctx, cred)
	if err != nil {
		return "", err
	}

	mixinKey := getMixinKey(imgKey + subKey)
	params["wts"] = fmt.Sprintf("%d", time.Now().Unix())

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(params[k])))
	}
	query := strings.Join(queryParts, "&")

	hash := md5.Sum([]byte(query + mixinKey))
	wRid := hex.EncodeToString(hash[:])

	return query + "&w_rid=" + wRid, nil
}

// CallWbi wraps WBI signature before Call
func (c *APIClient) CallWbi(ctx context.Context, action string, req *resty.Request, method, baseUrl string, params map[string]string, cred *auth.Credential) ([]byte, error) {
	queryString, err := c.SignWbi(ctx, cred, params)
	if err != nil {
		return nil, NewError(ErrCodeUpstreamError, fmt.Sprintf("%s: failed to sign WBI params: %v", action, err))
	}
	fullUrl := baseUrl + "?" + queryString
	return c.Call(ctx, action, req, method, fullUrl)
}
