package riders

import "context"

type service struct {
	repo *repo
}

func newService(repo *repo) *service {
	return &service{repo: repo}
}

func (s *service) GetAllRiders(ctx context.Context) ([]Rider, error) {

	riders, err := s.repo.GetAllRiders(ctx)
	if err != nil {
		return nil, err
	}
	return riders, nil
}

func (s *service) GetRiderById(ctx context.Context, id int) (*Rider, error) {
	return s.repo.GetRiderDetails(ctx, id)
}
