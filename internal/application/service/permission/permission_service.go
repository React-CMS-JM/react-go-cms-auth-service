// Package permission applies permission queries.
package permission

// Service coordinates permission queries.
type Service struct {
	repository Repository
}

// New builds a permission service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
