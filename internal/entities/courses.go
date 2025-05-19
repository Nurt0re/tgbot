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
	Step     string
	Selected *Course
}
type Message struct {
	Role      string
	Text      string
	Timestamp string
}
