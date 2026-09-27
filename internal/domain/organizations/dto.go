package organizations

type MyOrganizationDto struct {
	ID          int64    `json:"id" example:"1"`
	Name        string   `json:"name" example:"Acme Esports"`
	Description *string  `json:"description" example:"The Acmes competitive division"`
	Logo        *string  `json:"logo" extensions:"x-nullable"`
	Banner      *string  `json:"banner" extensions:"x-nullable"`
	Active      bool     `json:"active" example:"true"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type OrganizationRoleGroupDto struct {
	Role          string              `json:"role" example:"OWNER"`
	Organizations []MyOrganizationDto `json:"organizations"`
}
