package teams

import "arena/internal/domain/users"

type TeamReadModel struct {
	Team    Team
	Country users.Country
	Owner   *users.User
}
