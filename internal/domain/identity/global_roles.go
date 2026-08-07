package identity

func HasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func IsSuperAdmin(roles []string) bool {
	return HasRole(roles, RoleSuperAdmin)
}

func IsStaff(roles []string) bool {
	return HasRole(roles, RoleStaff)
}

func IsUser(roles []string) bool {
	return HasRole(roles, RoleUser)
}

func IsStaffOrSuperAdmin(roles []string) bool {
	return IsStaff(roles) || IsSuperAdmin(roles)
}

func IsBanned(roles []string) bool {
	return HasRole(roles, RoleBanned)
}
