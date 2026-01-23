package rules

import grules "arena/pkg/global_rules"

type TeamActionContext struct {
	Actor        grules.ActorContext
	TargetUserID int64
	TargetRole   string
}

type TeamGameContext struct {
	Actor  grules.ActorContext
	GameId int32
}
