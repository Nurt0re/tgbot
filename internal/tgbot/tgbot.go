package tgbot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"tgbot/internal/entities"
	"tgbot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TestQuestion struct {
	Question string
	Options  []string
	Answer   int
}

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

		HandleConversation(db, bot, courses, update)
	}

	return nil
}

var userStates = make(map[int64]*entities.UserState)

func HandleConversation(botDB *sql.DB, bot *tgbotapi.BotAPI, courses []entities.Course, update tgbotapi.Update) {
	text := update.Message.Text
	chatID := update.Message.Chat.ID
	userState := getUserState(chatID)

	if userState.Step == "" && userState.Name == "" {
		userState.Step = "waiting_for_name"
		msg := tgbotapi.NewMessage(chatID, "Привет! Напиши своё имя, чтобы начать.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке приветствия: %v", err)
		}
		return
	}

	if userState.Step == "waiting_for_name" {
		userState.Name = strings.TrimSpace(text)
		userState.Step = "waiting_for_phone"
		msg := tgbotapi.NewMessage(chatID, "Спасибо! Теперь, пожалуйста, напиши свой номер телефона.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке запроса номера телефона: %v", err)
		}
		return
	}

	if userState.Step == "waiting_for_phone" {
		userState.PhoneNumber = strings.TrimSpace(text)
		userState.Step = "taking_test"
		userState.IsTakingTest = true
		userState.TestIndex = 0
		userState.TestScore = 0

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Спасибо, %s! Сейчас начнётся тест из 10 вопросов. Отвечай, выбрав номер варианта от 1 до 4.", userState.Name)))
		sendTestQuestion(bot, chatID, 0)
		return
	}

	if userState.IsTakingTest {
		handleTestStep(bot, botDB, chatID, text, userState, courses)
		return
	}

	switch text {
	case "/history":
		sendHistory(botDB, bot, chatID)
		return
	case "/courses":
		sendCourses(bot, chatID, courses)
		return
	case "/teachers":
		sendTeachers(bot, chatID, courses)
		return
	case "/schedule":
		sendSchedules(bot, chatID, courses)
		return
	case "/enrollments":
		sendEnrollments(botDB, bot, chatID)
		return

	}

	// Сохраняем входящее сообщение
	if err := storage.SaveMessage(botDB, chatID, "user", text); err != nil {
		log.Printf("Ошибка при сохранении входящего сообщения: %v", err)
	}

	var msg tgbotapi.MessageConfig

	switch userState.Step {
	case "":
		msg = handleInitialStep(chatID, text, userState, courses)

	case "waiting_for_course_selection":
		msg = handleCourseSelection(chatID, text, userState, courses)

	case "waiting_for_payment_confirmation":
		msg = handlePaymentConfirmation(chatID, text, userState, bot, botDB)

	case "payment_successful":
		msg = tgbotapi.NewMessage(chatID, "Напишите 'Выбрать курс' для начала.")
		userState.Step = ""

	default:
		msg = tgbotapi.NewMessage(chatID, "Неизвестный шаг. Напишите 'Выбрать курс' для начала.")
		userState.Step = ""
	}

	// Сохраняем исходящее сообщение
	if err := storage.SaveMessage(botDB, chatID, "bot", msg.Text); err != nil {
		log.Printf("Ошибка при сохранении исходящего сообщения: %v", err)
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func getUserState(chatID int64) *entities.UserState {
	if state, exists := userStates[chatID]; exists {
		return state
	}
	userStates[chatID] = &entities.UserState{}
	return userStates[chatID]
}

func sendEnrollments(db *sql.DB, bot *tgbotapi.BotAPI, chatID int64) {
	enrollments, err := storage.GetAllEnrollments(db)
	if err != nil {
		log.Printf("Ошибка при получении записей: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при получении записей."))
		return
	}

	if len(enrollments) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Записей пока нет."))
		return
	}

	var b strings.Builder
	b.WriteString("📋 Список записей:\n\n")
	for i, e := range enrollments {
		status := "❌ Не оплачено"
		if e.IsPaid {
			status = "✅ Оплачено"
		}
		b.WriteString(fmt.Sprintf("%d) %s (%s) — %s — %s — %s\n", i+1, e.Name, e.PhoneNumber, e.CourseName, status, e.Timestamp))
	}

	// Разбивка на части, если слишком длинно
	const chunkSize = 4000
	text := b.String()
	for len(text) > 0 {
		end := chunkSize
		if len(text) < chunkSize {
			end = len(text)
		}
		bot.Send(tgbotapi.NewMessage(chatID, text[:end]))
		text = text[end:]
	}
}

func handleInitialStep(chatID int64, text string, state *entities.UserState, courses []entities.Course) tgbotapi.MessageConfig {
	if state.Name == "" {
		state.Name = text
		state.Step = "name_entered"
		return tgbotapi.NewMessage(chatID, "Спасибо! Теперь напишите 'Выбрать курс', чтобы выбрать курс.")
	}
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

func handlePaymentConfirmation(chatID int64, text string, state *entities.UserState, bot *tgbotapi.BotAPI, botDB *sql.DB) tgbotapi.MessageConfig {
	if strings.EqualFold(text, "Да") {
		state.Step = "payment_successful"
		storage.SaveEnrollment(botDB, chatID, state.Name, state.Selected.Name, true, state.PhoneNumber)
		return tgbotapi.NewMessage(chatID, "Отлично! Ваш платеж был успешно принят. Спасибо за оплату!")
	} else if strings.EqualFold(text, "Нет") {
		state.Step = ""
		storage.SaveEnrollment(botDB, chatID, state.Name, state.Selected.Name, false, state.PhoneNumber)
		go remindUserLater(bot, botDB, chatID, state.Selected.Name, 7*24*time.Hour) // Pass botDB and courseName
		return tgbotapi.NewMessage(chatID, "Хорошо, подумайте еще. Напишите 'Выбрать курс', чтобы изменить выбор.")
	}

	return tgbotapi.NewMessage(chatID, "Пожалуйста, напишите 'Да' если вы оплатили, или 'Нет' если еще не оплатили.")
}

func parseCourseSelection(input string) (int, error) {
	var number int
	_, err := fmt.Sscanf(input, "%d", &number)
	return number, err
}

func remindUserLater(bot *tgbotapi.BotAPI, db *sql.DB, chatID int64, courseName string, duration time.Duration) { // Added db and courseName parameters
	time.Sleep(duration)

	userState := getUserState(chatID)
	// Проверяем, оплатил ли пользователь курс за это время
	enrollments, err := storage.GetEnrollmentsByUserIDAndCourse(db, chatID, courseName) // Use passed db and courseName
	if err == nil && len(enrollments) > 0 && enrollments[0].IsPaid {
		return // Пользователь уже оплатил
	}

	// Если курс все еще не оплачен или пользователь не выбрал новый
	if userState.Selected != nil && userState.Selected.Name == courseName { // Check if the reminder is still for the same course
		msgText := fmt.Sprintf("Напоминаем, что вы выбрали курс '%s', но еще не оплатили его. Напишите 'Выбрать курс', чтобы выбрать другой курс, или свяжитесь с нами для оплаты.", courseName)
		msg := tgbotapi.NewMessage(chatID, msgText)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending reminder: %v", err)
		}
	} else if userState.Step == "" || userState.Step == "waiting_for_course_selection" { // Generic reminder if no specific course context or user moved on
		msg := tgbotapi.NewMessage(chatID, "Вы еще не выбрали курс или не завершили оплату. Напишите 'Выбрать курс' чтобы начать заново.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending reminder: %v", err)
		}
	}
}

func sendHistory(db *sql.DB, bot *tgbotapi.BotAPI, chatID int64) {
	history, err := storage.GetConversationHistory(db, chatID)
	if err != nil {
		log.Printf("Ошибка при получении истории: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при получении истории."))
		return
	}

	if len(history) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Диалог пуст."))
		return
	}

	var builder strings.Builder
	for _, msg := range history {
		builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", msg.Timestamp, msg.Role, msg.Text))
	}

	const chunkSize = 4000
	text := builder.String()
	for len(text) > 0 {
		end := chunkSize
		if len(text) < chunkSize {
			end = len(text)
		}
		bot.Send(tgbotapi.NewMessage(chatID, text[:end]))
		text = text[end:]
	}
}

func sendCourses(bot *tgbotapi.BotAPI, chatID int64, courses []entities.Course) {
	if len(courses) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Курсы отсутствуют."))
		return
	}

	var b strings.Builder
	b.WriteString("📚 Доступные курсы:\n\n")
	for i, course := range courses {
		b.WriteString(fmt.Sprintf("%d) %s — %.2f₽\n", i+1, course.Name, course.Price))
	}
	bot.Send(tgbotapi.NewMessage(chatID, b.String()))
}

func sendTeachers(bot *tgbotapi.BotAPI, chatID int64, courses []entities.Course) {
	teachersMap := make(map[string]bool)
	var b strings.Builder
	b.WriteString("👨‍🏫 Преподаватели:\n\n")
	for _, course := range courses {
		if !teachersMap[course.Teacher] {
			b.WriteString(fmt.Sprintf("- %s\n", course.Teacher))
			teachersMap[course.Teacher] = true
		}
	}
	bot.Send(tgbotapi.NewMessage(chatID, b.String()))
}

func sendSchedules(bot *tgbotapi.BotAPI, chatID int64, courses []entities.Course) {
	if len(courses) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Расписание отсутствует."))
		return
	}

	var b strings.Builder
	b.WriteString("🗓 Расписание курсов:\n\n")
	for _, course := range courses {
		b.WriteString(fmt.Sprintf("- %s: %s\n", course.Name, course.Schedule))
	}
	bot.Send(tgbotapi.NewMessage(chatID, b.String()))
}

func sendTestQuestion(bot *tgbotapi.BotAPI, chatID int64, index int) {
	q := testQuestions[index]
	var b strings.Builder
	b.WriteString(fmt.Sprintf("❓ %s\n", q.Question))
	for i, option := range q.Options {
		b.WriteString(fmt.Sprintf("%d) %s\n", i+1, option))
	}
	bot.Send(tgbotapi.NewMessage(chatID, b.String()))
}

func handleTestStep(bot *tgbotapi.BotAPI, db *sql.DB, chatID int64, userAnswer string, state *entities.UserState, courses []entities.Course) {
	answerIndex := -1
	_, err := fmt.Sscanf(userAnswer, "%d", &answerIndex)

	// Проверка: введено не число или не от 1 до 4
	if err != nil || answerIndex < 1 || answerIndex > 4 {
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите число от 1 до 4.")
		bot.Send(msg)
		// Повторить текущий вопрос
		sendTestQuestion(bot, chatID, state.TestIndex)
		return
	}

	// Проверка ответа
	if answerIndex-1 == testQuestions[state.TestIndex].Answer {
		state.TestScore++
	}

	state.TestIndex++

	if state.TestIndex >= len(testQuestions) {
		state.IsTakingTest = false
		state.Step = ""
		// Сохраняем результат теста
		if err := storage.SaveTestResult(db, chatID, state.Name, state.TestScore); err != nil {
			log.Printf("Ошибка при сохранении результата теста: %v", err)
		}
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Тест завершён! Вы набрали %d из %d баллов.", state.TestScore, len(testQuestions))))
		sendRecommendedCourses(bot, db, chatID, state.TestScore)

		// Немного подождать перед следующим шагом
		time.Sleep(2 * time.Second)

		// Перевести пользователя к выбору курса
		state.Step = "waiting_for_course_selection"
		msg := handleInitialStep(chatID, "Выбрать курс", state, courses)
		bot.Send(msg)
		return
	}

	// Следующий вопрос
	sendTestQuestion(bot, chatID, state.TestIndex)
}

func sendRecommendedCourses(bot *tgbotapi.BotAPI, db *sql.DB, chatID int64, score int) {
	var level string
	if score <= 3 {
		level = "Начальный"
	} else if score <= 7 {
		level = "Средний"
	} else {
		level = "Продвинутый"
	}

	courses, _ := storage.GetCourses(db) // если нужно — передай подключение

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📚 Курсы уровня \"%s\":\n\n", level))
	i := 1
	for _, course := range courses {
		if strings.EqualFold(course.Level, level) {
			b.WriteString(fmt.Sprintf("%d) %s — %.2f₽\n", i, course.Name, course.Price))
			i++
		}
	}

	if i == 1 {
		b.WriteString("Нет доступных курсов для этого уровня.")
	}

	bot.Send(tgbotapi.NewMessage(chatID, b.String()))
}
