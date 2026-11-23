package tournaments

type Module struct {
	service *Service
}

func NewModule(service *Service) *Module {
	return &Module{service: service}
}
