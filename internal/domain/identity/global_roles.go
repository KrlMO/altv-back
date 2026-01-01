package identity

func HasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func IsIluvatar(roles []string) bool {
	return HasRole(roles, RoleIluvatar)
}

func IsAinur(roles []string) bool {
	return HasRole(roles, RoleAinur)
}

func IsMaiar(roles []string) bool {
	return HasRole(roles, RoleMaiar)
}

func IsElf(roles []string) bool {
	return HasRole(roles, RoleElf)
}

func IsHuman(roles []string) bool {
	return HasRole(roles, RoleHuman)
}

func IsIluvatarOrAinur(roles []string) bool {
	return IsIluvatar(roles) || IsAinur(roles)
}

func IsOrc(roles []string) bool {
	return HasRole(roles, RoleOrc)
}
