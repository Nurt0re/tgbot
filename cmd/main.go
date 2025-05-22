package main

import (
	"fmt"
	"log"
	"os"
	"tgbot/internal/storage"
	"tgbot/internal/tgbot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	defer db.Close()
	storage.SeedCourses(db)

	token := os.Getenv("TGBOT_API")
	if token == "" {
		log.Fatal("Токен не задан в переменной окружения TGBOT_API")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot started...")

	if err := tgbot.Run(bot, db); err != nil {
		log.Fatal(err)
	}
}

// func parseCourseSelection(text string) (int, error) {
// 	var courseNumber int
// 	_, err := fmt.Sscanf(text, "%d", &courseNumber)
// 	return courseNumber, err
// }
