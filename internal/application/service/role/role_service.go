// Package role applies role queries.
package role

// Service coordinates role queries.
type Service struct {
	repository Repository
}

// New builds a role service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
