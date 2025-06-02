package gpt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkoukk/tiktoken-go"
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

// countTokens 計算 text 在指定 model 下的 token 數量
func countTokens(model string, text string, tokenCount *int) {
	// 根據 model 取得對應的 tokenizer
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		log.Printf("EncodingForModel error: %v, fallback to cl100k_base\n", err)
		// 如果錯誤就用預設的 cl100k_base 編碼
		enc, err = tiktoken.GetEncoding("cl100k_base")
	}

	// 編碼文字成 token IDs
	tokens := enc.Encode(text, nil, nil)

	*tokenCount += len(tokens) // 更新 tokenCount 的值
	// return len(tokens)
}

func (s *GptService) SendWithRetry(req GptContentRequest) (*openai.ChatCompletionStream, error) {

	RoleSystemContent := "你是一個編輯器助手，會按照需求回應html內容，所有的樣式會寫在style標籤中，並且會在body標籤中寫入內容，請不要回覆任何其他內容，只回覆html內容。不需要產生html標籤，最外層請直接使用div標籤。"
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: RoleSystemContent,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Content,
		},
		// {
		// 	Role:    openai.ChatMessageRoleAssistant,
		// 	Content: req.Content,
		// },
	}
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if err := s.limiter.Wait(context.Background()); err != nil {
			return nil, err
		}

		stream, err := s.client.CreateChatCompletionStream(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:     openai.GPT4o,
				MaxTokens: 5000, // 限制回覆最大長度
				// Temperature: 0.7, // 可加上創造力參數
				Stream:   true, // 即時回覆
				Messages: messages,
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

	enc, _ := tiktoken.EncodingForModel("gpt-4o")
	tokens := enc.Encode("hello world", nil, nil)
	fmt.Println(tokens)      // 例如輸出: [15339 1917]
	fmt.Println(len(tokens)) // 輸出: 2（表示有兩個 tokens）

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
	// 累積 Assistant 回覆內容
	var fullReply string
	var tokenCount int
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// 串流結束
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

			// 記錄並計算 token
			fullReply += content
			go countTokens("gpt-4o", content, &tokenCount)
		}
	}
	fmt.Printf("tokenCount: %d\n", tokenCount)

	// return &stream.Choices[0].Message.Content, err
	return nil, nil
}
