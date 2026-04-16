package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractBvid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      bool
	}{
		{"BV1test12345", "BV1test12345", false},
		{"https://www.bilibili.com/video/BV1test12345", "BV1test12345", false},
		{"invalid", "", true},
		{"BV123", "", true},
	}

	for _, tt := range tests {
		result, err := ExtractBvid(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("Expected error for input %q, got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		}
	}
}

func TestMapAPIError(t *testing.T) {
	err := MapAPIError("Test", -101, "Not logged in")
	if bErr, ok := err.(*BiliError); ok {
		if bErr.Code != ErrCodeNotAuthenticated {
			t.Errorf("Expected code %q, got %q", ErrCodeNotAuthenticated, bErr.Code)
		}
	} else {
		t.Errorf("Expected BiliError")
	}

	err = MapAPIError("Test", -404, "Not found")
	if bErr, ok := err.(*BiliError); ok {
		if bErr.Code != ErrCodeNotFound {
			t.Errorf("Expected code %q, got %q", ErrCodeNotFound, bErr.Code)
		}
	}
}

func TestClientRateLimitRetry(t *testing.T) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests < 2 {
			w.WriteHeader(http.StatusPreconditionFailed) // 412
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0,"data":{"test":"ok"}}`))
	}))
	defer ts.Close()

	c := NewClient()
	// reduce retry wait for fast testing
	c.resty.SetRetryWaitTime(1)

	ctx := context.Background()
	_, err := c.Call(ctx, "Test", c.R(ctx, nil), "GET", ts.URL)

	if err != nil {
		t.Errorf("Expected success after retry, got error: %v", err)
	}

	if requests != 2 {
		t.Errorf("Expected 2 requests, got %d", requests)
	}
}
