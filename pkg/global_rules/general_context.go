package grules

type ActorContext struct {
	ActorID          int64
	ActorGlobalRoles []string
	ActorTeamRoles   []string
}
