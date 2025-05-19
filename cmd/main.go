package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

type Course struct {
	Name        string
	Level       string
	Teacher     string
	Schedule    string
	Description string
}

func main() {
	// Замените YOUR_BOT_API_KEY на ваш токен
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	token := os.Getenv("TGBOT_API")
	if token == "" {
		log.Fatal("Токен не задан в переменной окружения TGBOT_API")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot started...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// Список курсов
	courses := []Course{
		{
			Name:        "Go для начинающих",
			Level:       "Начальный",
			Teacher:     "Иван Иванов",
			Schedule:    "Понедельно, 18:00-20:00",
			Description: "Основы языка Go. Изучение синтаксиса и базовых структур данных.",
		},
		{
			Name:        "Go для продвинутых",
			Level:       "Продвинутый",
			Teacher:     "Алексей Петров",
			Schedule:    "Вторник и четверг, 19:00-21:00",
			Description: "Продвинутые техники работы с Go, асинхронное программирование, паттерны проектирования.",
		},
		{
			Name:        "Python для начинающих",
			Level:       "Начальный",
			Teacher:     "Мария Сидорова",
			Schedule:    "Среда, 17:00-19:00",
			Description: "Основы Python. Создание простых программ и работа с библиотеками.",
		},
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := strings.ToLower(update.Message.Text)

		var msg tgbotapi.MessageConfig

		// Обрабатываем команды
		switch {
		case strings.HasPrefix(text, "/courses"):
			coursesText := "Доступные курсы:\n\n"
			for _, course := range courses {
				coursesText += fmt.Sprintf(
					"Название: %s\nУровень: %s\nПреподаватель: %s\nВремя: %s\nОписание: %s\n\n",
					course.Name, course.Level, course.Teacher, course.Schedule, course.Description,
				)
			}
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, coursesText)

		case strings.HasPrefix(text, "/teachers"):
			teachersText := "Преподаватели:\n\n"
			for _, course := range courses {
				teachersText += fmt.Sprintf("Преподаватель: %s — %s\n", course.Teacher, course.Name)
			}
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, teachersText)

		case strings.HasPrefix(text, "/schedule"):
			scheduleText := "Расписание курсов:\n\n"
			for _, course := range courses {
				scheduleText += fmt.Sprintf("Курс: %s — Время: %s\n", course.Name, course.Schedule)
			}
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, scheduleText)

		default:
			if strings.Contains(text, "привет") {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я могу помочь выбрать курс, узнать расписание и преподавателей. Напиши /courses для списка курсов.")
			} else if strings.Contains(text, "как дела") {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "У меня все хорошо! Готов помочь выбрать курс. Напиши /courses.")
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Я не понял запрос. Напиши /courses для списка курсов или /schedule для расписания.")
			}
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
