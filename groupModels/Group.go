package groupModels

type LeaderBoardItem struct {
	Username string
	Score    float32
}

const USER_GROUP_LIMIT = 20

type Link struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Expire string `json:"expire"`
}

type Group struct {
	GroupName   string
	GroupID     string
	CreatedAt   string
	GroupTime   float32
	LeaderBoard []LeaderBoardItem
	Links       []Link
	CreatedBy   string
}

type JoinGroupModel struct {
	GroupID  string
	Username string
	Time     float64
}
