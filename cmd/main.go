package main

import (
	"fmt"
	"log"
	"os"
	"tgbot/internal/entities"
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

var userStates = make(map[int64]*entities.UserState)

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	defer db.Close()
	storage.SeedCourses(db)

	// Получаем токен из переменной окружения
	token := os.Getenv("TGBOT_API")
	if token == "" {
		log.Fatal("Токен не задан в переменной окружения TGBOT_API")
	}

	// Создаем нового бота с токеном из переменной окружения
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot started...")

	if err := tgbot.Run(bot, db); err != nil {
		log.Fatal(err)
	}
}

// Функция для парсинга выбора курса пользователя
func parseCourseSelection(text string) (int, error) {
	var courseNumber int
	_, err := fmt.Sscanf(text, "%d", &courseNumber)
	return courseNumber, err
}
