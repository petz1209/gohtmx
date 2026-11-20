package riders

import (
	"context"
	"database/sql"
	"fmt"
)

type repo struct {
	db *sql.DB
}

func newRepo(db *sql.DB) *repo {
	return &repo{db: db}
}

func (r *repo) GetAllRiders(ctx context.Context) ([]Rider, error) {
	stmt := `SELECT r.id, r.first_name, r.last_name, r.nickname, r.nationality, r.birth_date, r.discipline,
				r.team_id, t.name as team
				FROM riders r
				LEFT JOIN teams t ON r.team_id = t.id
	`
	if r.db == nil {
		return nil, fmt.Errorf("repo.db is nil")
	}
	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("query all riders failed: %w", err)
	}
	defer rows.Close()

	data := make([]Rider, 0)

	var stringBirthday *string

	for rows.Next() {
		var rider Rider
		if err := rows.Scan(
			&rider.ID,
			&rider.FirstName,
			&rider.LastName,
			&rider.NickName,
			&rider.Nationality,
			&stringBirthday,
			&rider.Discipline,
			&rider.TeamID,
			&rider.Team,
		); err != nil {
			return nil, fmt.Errorf("query all riders failed: %w", err)
		}
		data = append(data, rider)
	}
	return data, nil

}

func (r *repo) GetRiderDetails(ctx context.Context, id int) (*Rider, error) {

	stmt := `SELECT r.id, r.first_name, r.last_name, r.nickname, r.nationality, r.birth_date, r.discipline,
				r.team_id, t.name as team
				FROM riders r
				LEFT JOIN teams t ON r.team_id = t.id
				WHERE r.id = $1
				`

	var rider Rider

	var stringBirthDay *string

	if err := r.db.QueryRowContext(ctx, stmt, id).Scan(&rider.ID,
		&rider.FirstName,
		&rider.LastName,
		&rider.NickName,
		&rider.Nationality,
		&stringBirthDay,
		&rider.Discipline,
		&rider.TeamID,
		&rider.Team); err != nil {
		return nil, err
	}

	return &rider, nil
}
