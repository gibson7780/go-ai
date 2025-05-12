package gpt

type GptContentRequest struct {
	Content string `json:"content" binding:"required"`
}
