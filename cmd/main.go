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

<<<<<<< HEAD
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
=======
var bot *tgbotapi.BotAPI

type Course struct {
	Name  string
	Level string
	Price float64
}

var courses = []Course{
	{
		Name:  "Go для начинающих",
		Level: "Начальный",
		Price: 1000.0,
	},
	{
		Name:  "Go для продвинутых",
		Level: "Продвинутый",
		Price: 1500.0,
	},
	{
		Name:  "Python для начинающих",
		Level: "Начальный",
		Price: 1200.0,
	},
}

type UserState struct {
	Step     string
	Selected *Course
>>>>>>> be6e0c4872fbc43478c70a8a4b76133733d22f2b
}

var userStates = make(map[int64]*UserState)

func main() {
<<<<<<< HEAD
	db, err := storage.InitDB()
=======
	err := godotenv.Load(".env")
>>>>>>> be6e0c4872fbc43478c70a8a4b76133733d22f2b
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

<<<<<<< HEAD
	if err := tgbot.Run(bot, db); err != nil {
		log.Fatal(err)
=======
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// text := strings.ToLower(update.Message.Text)
		text := (update.Message.Text)

		var msg tgbotapi.MessageConfig

		// Получаем состояние пользователя или создаем новое, если оно отсутствует
		userState, exists := userStates[update.Message.Chat.ID]
		if !exists {
			userState = &UserState{}
			userStates[update.Message.Chat.ID] = userState
		}

		switch userState.Step {
		case "":
			// Шаг 1: Выбор курса
			if text == "Выбрать курс" {
				coursesText := "Выберите курс:\n"
				for i, course := range courses {
					coursesText += fmt.Sprintf("%d) %s — %s — %0.2f₽\n", i+1, course.Name, course.Level, course.Price)
				}
				coursesText += "\nОтправьте номер курса для выбора."
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, coursesText)
				userState.Step = "waiting_for_course_selection"
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Напишите 'Выбрать курс' чтобы выбрать курс.")
			}

		case "waiting_for_course_selection":
			// Шаг 2: Обработка выбора курса
			if courseNumber, err := parseCourseSelection(text); err == nil && courseNumber >= 1 && courseNumber <= len(courses) {
				selectedCourse := &courses[courseNumber-1]
				userState.Selected = selectedCourse

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вы выбрали курс: %s.\nЦена: %.2f₽\nХотите оплатить? Напишите 'Да' или 'Нет'.", selectedCourse.Name, selectedCourse.Price))
				userState.Step = "waiting_for_payment_confirmation"
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный номер курса. Пожалуйста, выберите курс по номеру.")
			}

		case "waiting_for_payment_confirmation":
			// Шаг 3: Обработка подтверждения оплаты
			if strings.Contains(text, "Да") {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Отлично! Ваш платеж был успешно принят. Спасибо за оплату!")
				userState.Step = "payment_successful"
			} else if strings.Contains(text, "Нет") {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Хорошо, подумайте еще. Напишите 'Выбрать курс', чтобы изменить выбор.")
				userState.Step = ""
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, напишите 'Да' если вы оплатили, или 'Нет' если еще не оплатили.")
			}

		case "payment_successful":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Напишите 'Выбрать курс' для начала.")
			userState.Step = ""

		default:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестный шаг. Напишите 'Выбрать курс' для начала.")
			userState.Step = ""
		}

		// Отправляем сообщение пользователю
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
>>>>>>> be6e0c4872fbc43478c70a8a4b76133733d22f2b
	}
}

// Функция для парсинга выбора курса пользователя
func parseCourseSelection(text string) (int, error) {
	var courseNumber int
	_, err := fmt.Sscanf(text, "%d", &courseNumber)
	return courseNumber, err
}
