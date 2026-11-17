package riders

import (
	"bikes/internal/features/riders/domain"
	"bikes/internal/features/riders/repository"
	"context"
)

type Service struct {
	repo *repository.Repo
}

func NewService(repo *repository.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRiders(ctx context.Context) ([]domain.Rider, error) {
	return s.repo.GetRiders(ctx)
}
