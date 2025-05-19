package tgbot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"tgbot/internal/entities"
	"tgbot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Run(bot *tgbotapi.BotAPI, db *sql.DB) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("failed to get updates: %w", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		handleUpdate(bot, db, update)
	}

	return nil
}

func handleUpdate(bot *tgbotapi.BotAPI, db *sql.DB, update tgbotapi.Update) {
	text := strings.ToLower(update.Message.Text)
	chatID := update.Message.Chat.ID

	courses, err := storage.GetCourses(db)
	if err != nil {
		log.Printf("DB error: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении курсов.")
		bot.Send(msg)
		return
	}

	var msg tgbotapi.MessageConfig

	switch {
	case strings.HasPrefix(text, "/courses"):
		msg = tgbotapi.NewMessage(chatID, formatCourses(courses))

	case strings.HasPrefix(text, "/teachers"):
		msg = tgbotapi.NewMessage(chatID, formatTeachers(courses))

	case strings.HasPrefix(text, "/schedule"):
		msg = tgbotapi.NewMessage(chatID, formatSchedule(courses))

	default:
		msg = tgbotapi.NewMessage(chatID, handleFallback(text))
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func formatCourses(courses []entities.Course) string {
	var b strings.Builder
	b.WriteString("Доступные курсы:\n\n")
	for _, course := range courses {
		b.WriteString(fmt.Sprintf(
			"Название: %s\nУровень: %s\nПреподаватель: %s\nВремя: %s\nОписание: %s\n\n",
			course.Name, course.Level, course.Teacher, course.Schedule, course.Description,
		))
	}
	return b.String()
}

func formatTeachers(courses []entities.Course) string {
	var b strings.Builder
	b.WriteString("Преподаватели:\n\n")
	for _, course := range courses {
		b.WriteString(fmt.Sprintf("Преподаватель: %s — %s\n", course.Teacher, course.Name))
	}
	return b.String()
}

func formatSchedule(courses []entities.Course) string {
	var b strings.Builder
	b.WriteString("Расписание курсов:\n\n")
	for _, course := range courses {
		b.WriteString(fmt.Sprintf("Курс: %s — Время: %s\n", course.Name, course.Schedule))
	}
	return b.String()
}

func handleFallback(text string) string {
	switch {
	case strings.Contains(text, "привет"):
		return "Привет! Я могу помочь выбрать курс. Напиши /courses."
	case strings.Contains(text, "как дела"):
		return "У меня все отлично, спасибо! Напиши /courses для списка курсов."
	default:
		return "Я не понял. Напиши /courses или /schedule."
	}
}
