package gpt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GptHandler struct {
	service *GptService
}

type Stream struct {
	Content string `json:"content"`
}

func NewGptHandler(service *GptService) *GptHandler {

	return &GptHandler{
		service: service,
	}
}

func (h *GptHandler) SendContentHandler(c *gin.Context) {

	var req GptContentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "無效的請求格式"})
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid request format: " + err.Error()})
		return
	}

	res, err := h.service.sendContent(c, req)

	if err != nil {
		c.JSON(400, gin.H{"error": "無效的請求格式"})
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid request format: " + err.Error()})
		return
	}

	if err != nil {
		fmt.Printf("呼叫 OpenAI API 出錯: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid request format: " + err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "successfully",
		"content": res,
	})
}

func (h *GptHandler) StreamHandler(c *gin.Context) {
	var req GptContentRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	// c.Writer.Header().Set("Content-Type", "text/plain")
	// c.Writer.Header().Set("Transfer-Encoding", "chunked")
	// c.Writer.Header().Set("Cache-Control", "no-cache")
	// c.Writer.Flush()
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// for i := 1; i <= 5; i++ {

	res, err := h.service.sendContent(c, req)
	fmt.Print(res)
	// _, err = c.Writer.Write([]byte(*res))
	if err != nil {
		log.Printf("Write error: %v", err)
	}
	// c.Writer.Flush()
	// for {
	// 	_, err := c.Writer.Write([]byte(*res))
	// 	if err != nil {
	// 		log.Printf("Write error: %v", err)
	// 	}
	// 	c.Writer.Flush()
	// 	// time.Sleep(1 * time.Second)
	// }

	// }
}
