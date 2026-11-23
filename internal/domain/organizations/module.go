package organizations

type Module struct {
	service       *Service
	createService *CreateOrganizationService
}

func NewModuleWithCreate(service *Service, createService *CreateOrganizationService) *Module {
	return &Module{service, createService}
}
