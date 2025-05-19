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

	courses, err := storage.GetCourses(db)
	if err != nil {
		return fmt.Errorf("failed to fetch courses: %w", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		HandleConversation(bot, courses, update)
	}

	return nil
}

var userStates = make(map[int64]*entities.UserState)

func HandleConversation(bot *tgbotapi.BotAPI, courses []entities.Course, update tgbotapi.Update) {
	text := update.Message.Text
	chatID := update.Message.Chat.ID
	userState := getUserState(chatID)

	var msg tgbotapi.MessageConfig

	switch userState.Step {
	case "":
		msg = handleInitialStep(chatID, text, userState, courses)

	case "waiting_for_course_selection":
		msg = handleCourseSelection(chatID, text, userState, courses)

	case "waiting_for_payment_confirmation":
		msg = handlePaymentConfirmation(chatID, text, userState)

	case "payment_successful":
		msg = tgbotapi.NewMessage(chatID, "Напишите 'Выбрать курс' для начала.")
		userState.Step = ""

	default:
		msg = tgbotapi.NewMessage(chatID, "Неизвестный шаг. Напишите 'Выбрать курс' для начала.")
		userState.Step = ""
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func getUserState(chatID int64) *entities.UserState {
	if state, exists := userStates[chatID]; exists {
		return state
	}
	userStates[chatID] = &entities.UserState{}
	return userStates[chatID]
}

func handleInitialStep(chatID int64, text string, state *entities.UserState, courses []entities.Course) tgbotapi.MessageConfig {
	if strings.EqualFold(text, "Выбрать курс") {
		var b strings.Builder
		b.WriteString("Выберите курс:\n\n")
		for i, course := range courses {
			b.WriteString(fmt.Sprintf(
				"%d) %s\nУровень: %s\nПреподаватель: %s\nВремя: %s\nОписание: %s\nЦена: %.2f₽\n\n",
				i+1, course.Name, course.Level, course.Teacher, course.Schedule, course.Description, course.Price,
			))
		}
		b.WriteString("Отправьте номер курса для выбора.")
		state.Step = "waiting_for_course_selection"
		return tgbotapi.NewMessage(chatID, b.String())
	}

	return tgbotapi.NewMessage(chatID, "Привет! Напишите 'Выбрать курс' чтобы выбрать курс.")
}

func handleCourseSelection(chatID int64, text string, state *entities.UserState, courses []entities.Course) tgbotapi.MessageConfig {
	courseNumber, err := parseCourseSelection(text)
	if err == nil && courseNumber >= 1 && courseNumber <= len(courses) {
		selectedCourse := &courses[courseNumber-1]
		state.Selected = selectedCourse
		state.Step = "waiting_for_payment_confirmation"
		return tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"Вы выбрали курс: %s.\nЦена: %.2f₽\nХотите оплатить? Напишите 'Да' или 'Нет'.",
			selectedCourse.Name, selectedCourse.Price,
		))
	}

	return tgbotapi.NewMessage(chatID, "Неверный номер курса. Пожалуйста, выберите курс по номеру.")
}

func handlePaymentConfirmation(chatID int64, text string, state *entities.UserState) tgbotapi.MessageConfig {
	if strings.EqualFold(text, "Да") {
		state.Step = "payment_successful"
		return tgbotapi.NewMessage(chatID, "Отлично! Ваш платеж был успешно принят. Спасибо за оплату!")
	} else if strings.EqualFold(text, "Нет") {
		state.Step = ""
		return tgbotapi.NewMessage(chatID, "Хорошо, подумайте еще. Напишите 'Выбрать курс', чтобы изменить выбор.")
	}

	return tgbotapi.NewMessage(chatID, "Пожалуйста, напишите 'Да' если вы оплатили, или 'Нет' если еще не оплатили.")
}

func parseCourseSelection(input string) (int, error) {
	var number int
	_, err := fmt.Sscanf(input, "%d", &number)
	return number, err
}
