package users

func mapAuth(auth []ProviderAuth) []ProvidersAuthResponse {
	var authResponses []ProvidersAuthResponse
	for _, a := range auth {
		authResponses = append(authResponses, ProvidersAuthResponse{
			ID:          a.ID,
			Provider:    *a.ProviderName,
			ProviderUID: *a.ProviderUID,
			IsVerified:  a.IsVerified,
		})
	}
	return authResponses
}
