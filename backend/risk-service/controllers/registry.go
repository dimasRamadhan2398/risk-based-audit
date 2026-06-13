package controllers

type Registry struct{}

func NewControllerRegistry() *Registry {
	return &Registry{}
}
