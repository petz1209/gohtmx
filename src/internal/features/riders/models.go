package riders

import "time"

/*
stmt := `SELECT r.id, r.first_name, r.last_name, r.nickname, r.nationality, r.birth_date, r.discipline,
				r.team_id, t.name as team
				FROM riders r
				LEFT JOIN teams t ON r.team_id = t.id

*/
type Rider struct {
	ID          int
	FirstName   string
	LastName    string
	NickName    *string
	Nationality *string
	BirthDate   *time.Time
	Discipline  *string
	TeamID      *int
	Team        *string
}
