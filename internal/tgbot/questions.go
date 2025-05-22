package tgbot

var testQuestions = []TestQuestion{
	{
		Question: "Что такое переменная в программировании?",
		Options:  []string{"Константа", "Указатель", "Область памяти с именем", "Цикл"},
		Answer:   2,
	},
	{
		Question: "Какой тип данных используется для целых чисел в Go?",
		Options:  []string{"float", "string", "bool", "int"},
		Answer:   3,
	},
	{
		Question: "Какой символ используется для начала комментария в Go?",
		Options:  []string{"//", "#", "--", "/*"},
		Answer:   0,
	},
	{
		Question: "Как объявить функцию в Go?",
		Options:  []string{"def", "function", "func", "fn"},
		Answer:   2,
	},
	{
		Question: "Какой ключ используется для условного оператора?",
		Options:  []string{"case", "for", "switch", "if"},
		Answer:   3,
	},
	{
		Question: "Как создать срез в Go?",
		Options:  []string{"array()", "[]", "slice{}", "{}"},
		Answer:   1,
	},
	{
		Question: "Что такое goroutine?",
		Options:  []string{"Тип данных", "Функция", "Отдельный поток выполнения", "Модуль"},
		Answer:   2,
	},
	{
		Question: "Как обозначается цикл с 5 итерациями?",
		Options:  []string{"repeat 5", "for i := 0; i < 5; i++", "loop 5", "foreach 5"},
		Answer:   1,
	},
	{
		Question: "Какой оператор используется для присваивания?",
		Options:  []string{"==", "->", "=", ":="},
		Answer:   2,
	},
	{
		Question: "Как обозначается пакет в начале файла Go?",
		Options:  []string{"import", "package", "main", "module"},
		Answer:   1,
	},
}
