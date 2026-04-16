package auth

// Credential represents a user's authentication details.
type Credential struct {
	SESSDATA      string `json:"sessdata"`
	BiliJCT       string `json:"bili_jct"`
	AcTimeValue   string `json:"ac_time_value,omitempty"`
	Buvid3        string `json:"buvid3,omitempty"`
	Buvid4        string `json:"buvid4,omitempty"`
	DedeUserID    string `json:"dedeuserid,omitempty"`
	SavedAt       int64  `json:"saved_at,omitempty"`
}
