package domain

type Rider struct {
	ID         int
	FirstName  string
	LastName   string
	Discipline string
}

type RiderDetails struct {
	ID          int
	FirstName   string
	LastName    string
	Discipline  string
	Nickname    string
	Nationality string
	BirthDate   string
}
