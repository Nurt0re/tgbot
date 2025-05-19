package storage

import (
	"database/sql"
	"tgbot/internal/entities"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./internal/storage/courses.db")
	if err != nil {
		return nil, err
	}

	if err := initTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func initTables(db *sql.DB) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			level TEXT,
			teacher TEXT,
			schedule TEXT,
			description TEXT,
			price REAL
		);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			role TEXT,
			message TEXT,
			timestamp DATETIME
		);`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func SeedCourses(db *sql.DB) {
	rows, _ := db.Query("SELECT id FROM courses LIMIT 1")
	defer rows.Close()
	if rows.Next() {
		return
	}

	courses := []entities.Course{
		{
			Name:        "Go для начинающих",
			Level:       "Начальный",
			Teacher:     "Иван Иванов",
			Schedule:    "Понедельно, 18:00-20:00",
			Description: "Основы языка Go. Изучение синтаксиса и базовых структур данных.",
			Price:       9900.00,
		},
		{
			Name:        "Go для продвинутых",
			Level:       "Продвинутый",
			Teacher:     "Алексей Петров",
			Schedule:    "Вторник и четверг, 19:00-21:00",
			Description: "Продвинутые техники работы с Go, асинхронное программирование, паттерны проектирования.",
			Price:       14900.00,
		},
		{
			Name:        "Python для начинающих",
			Level:       "Начальный",
			Teacher:     "Мария Сидорова",
			Schedule:    "Среда, 17:00-19:00",
			Description: "Основы Python. Создание простых программ и работа с библиотеками.",
			Price:       8900.00,
		},
	}

	stmt, _ := db.Prepare("INSERT INTO courses(name, level, teacher, schedule, description, price) VALUES (?, ?, ?, ?, ?, ?)")
	defer stmt.Close()

	for _, course := range courses {
		stmt.Exec(course.Name, course.Level, course.Teacher, course.Schedule, course.Description, course.Price)
	}
}

func GetCourses(db *sql.DB) ([]entities.Course, error) {
	rows, err := db.Query("SELECT name, level, teacher, schedule, description, price FROM courses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []entities.Course
	for rows.Next() {
		var c entities.Course
		if err := rows.Scan(&c.Name, &c.Level, &c.Teacher, &c.Schedule, &c.Description, &c.Price); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, nil
}

func SaveMessage(db *sql.DB, userID int64, role, message string) error {
	query := `
	INSERT INTO conversations(user_id, role, message, timestamp)
	VALUES (?, ?, ?, ?);`
	_, err := db.Exec(query, userID, role, message, time.Now())
	return err
}
func GetConversationHistory(db *sql.DB, userID int64) ([]entities.Message, error) {
	query := `SELECT role, message, timestamp FROM conversations WHERE user_id = ? ORDER BY timestamp ASC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entities.Message
	for rows.Next() {
		var msg entities.Message
		var ts time.Time
		if err := rows.Scan(&msg.Role, &msg.Text, &ts); err != nil {
			return nil, err
		}
		msg.Timestamp = ts.Format("2006-01-02 15:04:05")
		history = append(history, msg)
	}
	return history, nil
}
