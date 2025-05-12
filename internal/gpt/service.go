package gpt

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type GptService struct {
	client *openai.Client
}

func NewGptService() *GptService {

	aipKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(aipKey)

	return &GptService{
		client: client,
	}
}

func (s *GptService) sendContent(c *gin.Context, req GptContentRequest) (*string, error) {

	// fmt.Print("req.Content", req.Content)
	stream, err := s.client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4o,
			MaxTokens: 2000, // 限制回覆最大長度
			// Temperature: 0.7, // 可加上創造力參數
			Stream: true, // 即時回覆
			Messages: []openai.ChatCompletionMessage{
				// {
				// 	Role:    openai.ChatMessageRoleSystem,
				// 	Content: req.Content,
				// },
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.Content,
				},
				// {
				// 	Role:    openai.ChatMessageRoleAssistant,
				// 	Content: req.Content,
				// },
			},
		},
	)
	if err != nil {
		log.Fatalf("Error while getting completion: %v", err)
	}
	defer stream.Close()
	// Stream loop
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("Stream receive error: %v\n", err)
			break
		}

		if &response.Choices[0].Delta.Content != nil {
			content := response.Choices[0].Delta.Content
			c.Writer.Write([]byte(content))
			c.Writer.Flush()
		}
	}
	// return &stream.Choices[0].Message.Content, err
	return nil, nil
}
