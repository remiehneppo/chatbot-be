package types

type ChatRequest struct {
	ChatId   string    `json:"chat_id"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	ChatId  string   `json:"chat_id"`
	Message *Message `json:"message"`
}
