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

func (service *Service) Create(ctx context.Context, requester IdentitySummary, input CreateRequest) (Ticket, error) {
	if service == nil || service.repository == nil {
		return Ticket{}, errTicketServiceNotConfigured
	}
	return service.repository.Create(ctx, requester, input)
}
