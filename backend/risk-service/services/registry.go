package services

type Registry struct{}

func NewServiceRegistry() *Registry {
	return &Registry{}
}
