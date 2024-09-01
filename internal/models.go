package internal

type Message struct {
	Id       int    `json:"id"`
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
}
