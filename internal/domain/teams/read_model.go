package teams

import (
	"arena/internal/domain/users"
	"arena/pkg/model"
)

type TeamReadModel struct {
	Team    Team
	Country model.Country
	Owner   *users.User
}
