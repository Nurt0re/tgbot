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
	Name       string
	CourseName string
	IsPaid     bool
	Timestamp  string
}
