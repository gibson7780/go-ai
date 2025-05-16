package gpt

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/time/rate"
)

type GptService struct {
	client  *openai.Client
	limiter *rate.Limiter
}

func NewGptService() *GptService {

	aipKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(aipKey)
	// 每秒 0.2次請求
	// 這裡的 0.2 是每秒的請求速率，1 是突發請求的上限
	// 這意味著在任何給定的時間內，最多可以有 1 次請求被排隊
	// 這樣可以防止過多的請求導致 API 限制
	limiter := rate.NewLimiter(rate.Limit(0.2), 1)
	return &GptService{
		client:  client,
		limiter: limiter,
	}
}

func (s *GptService) SendWithRetry(req GptContentRequest) (*openai.ChatCompletionStream, error) {

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if err := s.limiter.Wait(context.Background()); err != nil {
			return nil, err
		}
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
			return nil, err
		} else {
			return stream, nil
		}
	}
	return nil, nil
}

func (s *GptService) SendContent(c *gin.Context, req GptContentRequest) (*string, error) {

	// 限制段時間一次數  並等待下一次token取得
	// if err := s.limiter.Wait(context.Background()); err != nil {
	// 	log.Printf("Rate limit exceeded: %v\n", err)
	// 	c.JSON(429, gin.H{"error": "Rate limit exceeded"})
	// 	return nil, err
	// }
	stream, err := s.SendWithRetry(req)
	if err != nil {
		// log.Printf("Error while getting completion: %v", err)
		return nil, err
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

			return nil, err
		}
		content := response.Choices[0].Delta.Content
		if content != "" {
			c.Writer.Write([]byte(content))
			c.Writer.Flush()
		}
	}

	// return &stream.Choices[0].Message.Content, err
	return nil, nil
}
