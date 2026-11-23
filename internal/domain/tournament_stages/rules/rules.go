package rules

import (
	"errors"
	"fmt"

	"arena/internal/domain/identity"
)

var (
	ErrBadIdStage          = errors.New("the stage does not exist or the id is incorrect")
	ErrUnknownFormat       = errors.New("unknown or inactive stage format")
	ErrNoPermission        = errors.New("user does not have permission to manage this tournament's stages")
	ErrStageNotPending     = errors.New("only pending stages can be modified or deleted")
	ErrInvalidStatusChange = errors.New("invalid stage status transition")
	ErrPreviousStageOpen   = errors.New("the previous stage is not completed yet")
	ErrInvalidConfig       = errors.New("invalid stage config")
)

// Stage format codes seeded in al_stage_format. New formats are added to the
// catalog with an INSERT and get their own config validator here.
const (
	FormatSingleElimination = "SINGLE_ELIMINATION"
	FormatDoubleElimination = "DOUBLE_ELIMINATION"
	FormatSwiss             = "SWISS"
	FormatGroupStage        = "GROUP_STAGE"
)

// configSpec declares the allowed config keys and their bounds for a format.
// Common keys are allowed for every format.
type configKey struct {
	min int
	max int // 0 means unbounded
}

var commonConfigKeys = map[string]configKey{
	"best_of":     {1, 9},
	"teams_count": {2, 0},
	"advancing":   {1, 0},
}

var formatConfigKeys = map[string]map[string]configKey{
	FormatSwiss: {
		"rounds":             {1, 20},
		"qualification_wins": {1, 20},
		"elimination_losses": {1, 20},
	},
	FormatGroupStage: {
		"groups":            {1, 32},
		"qualify_per_group": {1, 32},
		// Future door: "group_double_elimination": true, "group_format": "GSL"...
	},
	FormatSingleElimination: {
		"byes": {0, 16},
	},
	FormatDoubleElimination: {
		"byes": {0, 16},
	},
}

// ValidateConfig validates a stage config against its format code.
// Unknown keys are rejected so the UI never has to render config it
// does not understand.
func ValidateConfig(formatCode string, cfg map[string]interface{}) error {
	spec := formatConfigKeys[formatCode]
	if spec == nil {
		return fmt.Errorf("%w: no config spec for format %s", ErrInvalidConfig, formatCode)
	}

	allowed := make(map[string]configKey, len(commonConfigKeys)+len(spec))
	for k, v := range commonConfigKeys {
		allowed[k] = v
	}
	for k, v := range spec {
		allowed[k] = v
	}

	for key, value := range cfg {
		bounds, ok := allowed[key]
		if !ok {
			return fmt.Errorf("%w: key %q is not allowed for format %s", ErrInvalidConfig, key, formatCode)
		}
		num, ok := toFloat(value)
		if !ok {
			return fmt.Errorf("%w: key %q must be a number", ErrInvalidConfig, key)
		}
		if num != float64(int(num)) {
			return fmt.Errorf("%w: key %q must be an integer", ErrInvalidConfig, key)
		}
		if num < float64(bounds.min) {
			return fmt.Errorf("%w: key %q must be >= %d", ErrInvalidConfig, key, bounds.min)
		}
		if bounds.max > 0 && num > float64(bounds.max) {
			return fmt.Errorf("%w: key %q must be <= %d", ErrInvalidConfig, key, bounds.max)
		}
	}
	return nil
}

// toFloat normalizes JSON numbers (float64) and Go ints to float64.
func toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// CanManageStages decides whether the actor may create/update/delete stages
// of a tournament: organization OWNER/ADMIN, tournament
// TOURNAMENT_OWNER/TOURNAMENT_ADMIN, or global staff.
func CanManageStages(actorOrgRoles []string, actorTournamentRoles []string, actorGlobalRoles []string) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if identity.HasRole(actorOrgRoles, identity.RoleOrgOwner) ||
		identity.HasRole(actorOrgRoles, identity.RoleOrgAdmin) {
		return nil
	}
	if identity.HasRole(actorTournamentRoles, "TOURNAMENT_OWNER") ||
		identity.HasRole(actorTournamentRoles, "TOURNAMENT_ADMIN") {
		return nil
	}
	return ErrNoPermission
}

// CanTransitionStageStatus validates a status change for a stage.
// Allowed transitions: PENDING -> IN_PROGRESS -> COMPLETED.
func CanTransitionStageStatus(current, next string) error {
	switch {
	case current == "PENDING" && next == "IN_PROGRESS":
		return nil
	case current == "IN_PROGRESS" && next == "COMPLETED":
		return nil
	default:
		return ErrInvalidStatusChange
	}
}
