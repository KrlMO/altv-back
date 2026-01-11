package rules

type TeamActionContext struct {
	Actor        ActorContext
	TargetUserID int64
	TargetRole   string
}

type TeamGameContext struct {
	Actor  ActorContext
	GameId int32
}

type ActorContext struct {
	ActorID          int64
	ActorGlobalRoles []string
	ActorTeamRoles   []string
}
