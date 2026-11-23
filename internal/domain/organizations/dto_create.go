package organizations

// CreateOrganizationRequest is the payload to create an organization.
type CreateOrganizationRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}
