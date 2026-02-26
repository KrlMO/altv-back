package auth

import "arena/internal/domain/users"

func mapAuth(auth []users.ProviderAuth) []users.ProvidersAuthResponse {
	var authResponses []users.ProvidersAuthResponse
	for _, a := range auth {
		authResponses = append(authResponses, users.ProvidersAuthResponse{
			ID:          a.ID,
			Provider:    *a.ProviderName,
			ProviderUID: *a.ProviderUID,
			IsVerified:  a.IsVerified,
		})
	}
	return authResponses
}
