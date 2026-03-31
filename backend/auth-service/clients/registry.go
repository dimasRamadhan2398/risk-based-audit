package clients

type ClientRegistry struct{}

type IClientRegistry interface {
	
}

func NewClientRegistry() IClientRegistry {
	return &ClientRegistry{}
}