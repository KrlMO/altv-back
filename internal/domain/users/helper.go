package users

func mapAuth(auth []UserAuth) []AuthUserProfileResponse {
	var authResponses []AuthUserProfileResponse
	for _, a := range auth {
		authResponses = append(authResponses, AuthUserProfileResponse{
			ID:          a.ID,
			Provider:    *a.ProviderName,
			ProviderUID: *a.ProviderUID,
		})
	}
	return authResponses
}
