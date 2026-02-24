package auth

import "arena/internal/domain/users"

func mapAuth(auth []users.UserAuth) []AuthResponse {
	var authResponses []AuthResponse
	for _, a := range auth {
		authResponses = append(authResponses, AuthResponse{
			ID:          a.ID,
			Provider:    *a.ProviderName,
			ProviderUID: *a.ProviderUID,
		})
	}
	return authResponses
}
