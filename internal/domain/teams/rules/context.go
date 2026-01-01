package rules

type TeamActionContext struct {
	ActorID          int64
	ActorGlobalRoles []string
	ActorTeamRoles   []string

	TargetUserID int64
	TargetRole   string
}
