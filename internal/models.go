package internal

type Message struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
}
