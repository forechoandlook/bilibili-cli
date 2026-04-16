package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/forechoandlook/bilibili-cli/internal/auth"
)

const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

type APIClient struct {
	resty *resty.Client
}

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func NewClient() *APIClient {
	r := resty.New()
	r.SetTimeout(10 * time.Second)
	r.SetHeader("User-Agent", DefaultUserAgent)

	// Basic retry for 412 / network errors
	r.SetRetryCount(3)
	r.SetRetryWaitTime(1 * time.Second)
	r.SetRetryMaxWaitTime(5 * time.Second)
	r.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			return true
		}
		if r.StatusCode() == 412 || r.StatusCode() >= 500 {
			return true
		}
		// If response is JSON, check the inner code
		var apiResp APIResponse
		if err := json.Unmarshal(r.Body(), &apiResp); err == nil {
			if apiResp.Code == -412 || apiResp.Code == 412 {
				return true
			}
		}
		return false
	})

	return &APIClient{
		resty: r,
	}
}

func (c *APIClient) R(ctx context.Context, cred *auth.Credential) *resty.Request {
	req := c.resty.R().SetContext(ctx)
	if cred != nil {
		var cookies []string
		if cred.SESSDATA != "" {
			cookies = append(cookies, "SESSDATA="+cred.SESSDATA)
		}
		if cred.BiliJCT != "" {
			cookies = append(cookies, "bili_jct="+cred.BiliJCT)
		}
		if len(cookies) > 0 {
			req.SetHeader("Cookie", strings.Join(cookies, "; "))
		}
		if cred.BiliJCT != "" {
			req.SetHeader("csrf", cred.BiliJCT)
		}
	}
	return req
}

// Call wraps the request execution and error mapping
func (c *APIClient) Call(ctx context.Context, action string, req *resty.Request, method, url string) ([]byte, error) {
	resp, err := req.Execute(method, url)
	if err != nil {
		return nil, NewError(ErrCodeNetworkError, fmt.Sprintf("%s: %v", action, err))
	}

	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusPreconditionFailed { // 412
			return nil, NewError(ErrCodeRateLimited, fmt.Sprintf("%s: HTTP 412 Bilibili API rate limit", action))
		}
		return nil, NewError(ErrCodeNetworkError, fmt.Sprintf("%s: HTTP %d", action, resp.StatusCode()))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		// Sometimes it's just raw data or not json (e.g. subtitles)
		return resp.Body(), nil
	}

	if apiResp.Code != 0 {
		return nil, MapAPIError(action, apiResp.Code, apiResp.Message)
	}

	return apiResp.Data, nil
}

// Helper: BV extraction
var bvidRe = regexp.MustCompile(`\bBV[0-9A-Za-z]{10}\b`)

func ExtractBvid(urlOrBvid string) (string, error) {
	match := bvidRe.FindString(urlOrBvid)
	if match != "" {
		return match, nil
	}
	return "", NewError(ErrCodeInvalidInput, fmt.Sprintf("无法提取 BV 号: %s", urlOrBvid))
}
