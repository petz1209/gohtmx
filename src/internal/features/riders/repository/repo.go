package repository

import (
	"bikes/internal/features/riders/domain"
	"context"
	"database/sql"
)

type Repo struct {
	DB *sql.DB
}

func NewRiderRepository(db *sql.DB) *Repo {

	return &Repo{DB: db}

}

func (r *Repo) GetRiders(ctx context.Context) ([]domain.Rider, error) {
	stmt := `SELECT id, first_name, last_name, discipline FROM riders`
	rows, err := r.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	riders := make([]domain.Rider, 0)
	for rows.Next() {
		var rider domain.Rider

		if err := rows.Scan(&rider.ID,
			&rider.FirstName,
			&rider.LastName,
			&rider.Discipline,
		); err != nil {
			return nil, err
		}
		riders = append(riders, rider)
	}

	return riders, nil

}
