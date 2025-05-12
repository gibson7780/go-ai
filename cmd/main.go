package main

import (
	"fmt"
	"go-openai/internal/router"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load() // 預設會從專案根目錄載入 .env
	if err != nil {
		log.Fatal("⚠️ 無法載入 .env 檔案")
	}

	r := router.SetupRouter()

	// 啟動伺服器
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("啟動伺服器失敗: %v\n", err)
	}
}
