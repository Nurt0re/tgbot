package storage

import (
	"database/sql"
	"tgbot/internal/entities"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./internal/storage/courses.db")
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS courses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		level TEXT,
		teacher TEXT,
		schedule TEXT,
		description TEXT,
		price REAL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return db, nil
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
