package tournaments

type Module struct {
	service   *Service
	jwtSecret string
}

func NewModuleWithAuth(service *Service, jwtSecret string) *Module {
	return &Module{service, jwtSecret}
}
