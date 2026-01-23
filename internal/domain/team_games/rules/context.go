package rules

import grules "arena/pkg/global_rules"

type TeamGameActionContext struct {
	Actor              grules.ActorContext
	ActorTeamGameRoles []string
	TeamGameID         int64
}
