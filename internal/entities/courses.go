package entities

type Course struct {
	Name        string
	Level       string
	Price       float64
	Teacher     string
	Schedule    string
	Description string
}

type UserState struct {
	Step         string
	Name         string
	PhoneNumber  string // Added PhoneNumber
	Selected     *Course
	TestIndex    int
	TestScore    int
	IsTakingTest bool
}

type Message struct {
	Role      string
	Text      string
	Timestamp string
}
type Enrollment struct {
	Name        string
	CourseName  string
	IsPaid      bool
	Timestamp   string
	TestScore   int    // Added TestScore
	PhoneNumber string // Added PhoneNumber
}

type UserQuestion struct {
	UserID       int64
	QuestionText string
	Timestamp    string
}

type UserQuestionDetail struct {
	Name         string
	PhoneNumber  string
	QuestionText string
	Timestamp    string
}
