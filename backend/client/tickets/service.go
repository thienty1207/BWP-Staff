package tickets

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) List(ctx context.Context, query ListQuery) (ListResponse, error) {
	if service == nil || service.repository == nil {
		return ListResponse{}, errTicketServiceNotConfigured
	}
	return service.repository.List(ctx, query)
}
