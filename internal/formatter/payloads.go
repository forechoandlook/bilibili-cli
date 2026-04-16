package formatter

import "fmt"

type AuthStatusData struct {
	Authenticated bool `json:"authenticated" yaml:"authenticated"`
	User *UserSummary `json:"user,omitempty" yaml:"user,omitempty"`
}

type UserSummary struct {
	ID       string `json:"id" yaml:"id"`
	Name     string `json:"name" yaml:"name"`
	Username string `json:"username" yaml:"username"`
	Level    int    `json:"level" yaml:"level"`
	Coins    int    `json:"coins" yaml:"coins"`
	Sign     string `json:"sign" yaml:"sign"`
	Vip      interface{} `json:"vip" yaml:"vip"`
}

type RelationData struct {
	Following int `json:"following" yaml:"following"`
	Follower  int `json:"follower" yaml:"follower"`
}

type WhoamiData struct {
	User     *UserSummary `json:"user" yaml:"user"`
	Relation *RelationData `json:"relation" yaml:"relation"`
}

type VideoSummary struct {
	ID               string `json:"id" yaml:"id"`
	Bvid             string `json:"bvid" yaml:"bvid"`
	Aid              int    `json:"aid" yaml:"aid"`
	Title            string `json:"title" yaml:"title"`
	Description      string `json:"description" yaml:"description"`
	DurationSeconds  int    `json:"duration_seconds" yaml:"duration_seconds"`
	Duration         string `json:"duration" yaml:"duration"`
	URL              string `json:"url" yaml:"url"`
	Owner            VideoOwner `json:"owner" yaml:"owner"`
	Stats            VideoStats `json:"stats" yaml:"stats"`
}

type VideoOwner struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

type VideoStats struct {
	View     int `json:"view" yaml:"view"`
	Danmaku  int `json:"danmaku" yaml:"danmaku"`
	Like     int `json:"like" yaml:"like"`
	Coin     int `json:"coin" yaml:"coin"`
	Favorite int `json:"favorite" yaml:"favorite"`
	Share    int `json:"share" yaml:"share"`
}

type SubtitleItem struct {
	From    float64 `json:"from" yaml:"from"`
	To      float64 `json:"to" yaml:"to"`
	Content string  `json:"content" yaml:"content"`
}

type SubtitlePayload struct {
	Available bool           `json:"available" yaml:"available"`
	Format    string         `json:"format" yaml:"format"`
	Text      string         `json:"text" yaml:"text"`
	Items     []SubtitleItem `json:"items" yaml:"items"`
}

type Comment struct {
	ID         string `json:"id" yaml:"id"`
	Author     VideoOwner `json:"author" yaml:"author"`
	Message    string `json:"message" yaml:"message"`
	Like       int    `json:"like" yaml:"like"`
	ReplyCount int    `json:"reply_count" yaml:"reply_count"`
}

type Warning struct {
	Code    string `json:"code" yaml:"code"`
	Message string `json:"message" yaml:"message"`
}

type VideoCommandPayload struct {
	Video      *VideoSummary   `json:"video" yaml:"video"`
	Subtitle   SubtitlePayload `json:"subtitle" yaml:"subtitle"`
	AiSummary  string          `json:"ai_summary" yaml:"ai_summary"`
	Comments   []Comment       `json:"comments" yaml:"comments"`
	Related    []VideoSummary  `json:"related" yaml:"related"`
	Warnings   []Warning       `json:"warnings" yaml:"warnings"`
}

type ListPayload struct {
	Items interface{} `json:"items" yaml:"items"`
}

type SearchUser struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Sign   string `json:"sign" yaml:"sign"`
	Fans   int    `json:"fans" yaml:"fans"`
	Videos int    `json:"videos" yaml:"videos"`
}

type SearchVideo struct {
	ID       string `json:"id" yaml:"id"`
	Bvid     string `json:"bvid" yaml:"bvid"`
	Title    string `json:"title" yaml:"title"`
	Author   string `json:"author" yaml:"author"`
	Play     int    `json:"play" yaml:"play"`
	Duration string `json:"duration" yaml:"duration"`
}

type FavoriteFolder struct {
	ID         int    `json:"id" yaml:"id"`
	Title      string `json:"title" yaml:"title"`
	MediaCount int    `json:"media_count" yaml:"media_count"`
}

type FavoriteMedia struct {
	ID              string `json:"id" yaml:"id"`
	Bvid            string `json:"bvid" yaml:"bvid"`
	Title           string `json:"title" yaml:"title"`
	DurationSeconds int    `json:"duration_seconds" yaml:"duration_seconds"`
	Duration        string `json:"duration" yaml:"duration"`
	Upper           struct {
		Name string `json:"name" yaml:"name"`
	} `json:"upper" yaml:"upper"`
}

type FollowingUser struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
	Sign string `json:"sign" yaml:"sign"`
}

type HistoryItem struct {
	ID       string `json:"id" yaml:"id"`
	Bvid     string `json:"bvid" yaml:"bvid"`
	Title    string `json:"title" yaml:"title"`
	Author   string `json:"author" yaml:"author"`
	ViewedAt string `json:"viewed_at" yaml:"viewed_at"`
}

type WatchLaterItem struct {
	ID              string `json:"id" yaml:"id"`
	Bvid            string `json:"bvid" yaml:"bvid"`
	Title           string `json:"title" yaml:"title"`
	Author          string `json:"author" yaml:"author"`
	DurationSeconds int    `json:"duration_seconds" yaml:"duration_seconds"`
	Duration        string `json:"duration" yaml:"duration"`
}

type DynamicItem struct {
	ID             string `json:"id" yaml:"id"`
	Author         struct {
		Name string `json:"name" yaml:"name"`
	} `json:"author" yaml:"author"`
	PublishedAt    string `json:"published_at" yaml:"published_at"`
	PublishedLabel string `json:"published_label" yaml:"published_label"`
	Title          string `json:"title" yaml:"title"`
	Text           string `json:"text" yaml:"text"`
	Stats          struct {
		Comment int `json:"comment" yaml:"comment"`
		Like    int `json:"like" yaml:"like"`
	} `json:"stats" yaml:"stats"`
}

type ActionResult struct {
	Success bool                   `json:"success" yaml:"success"`
	Action  string                 `json:"action" yaml:"action"`
	Result  map[string]interface{} `json:"result,omitempty" yaml:"result,omitempty"`
}

func ToInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	}
	return 0
}

func ToFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	}
	return 0.0
}
