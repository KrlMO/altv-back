package registrations

import (
	"context"
	"errors"
	"strconv"

	"arena/internal/domain/identity"
	"arena/internal/domain/tournament_registrations/rules"
	"arena/pkg/storage"
)

type Service struct {
	repo    *Repository
	storage storage.Storage
}

func NewService(repo *Repository, mediaStorage storage.Storage) *Service {
	return &Service{repo, mediaStorage}
}

// actorContext gathers the roles the actor holds on the team owning the
// roster, the roster itself, the tournament's organization and the tournament.
type actorContext struct {
	TeamRoles         []string
	TeamGameRoles     []string
	OrganizationRoles []string
	TournamentRoles   []string
}

func (s *Service) buildActorContext(
	ctx context.Context,
	actorID int64,
	teamID, teamGameID, organizationID, tournamentID int64,
) (*actorContext, error) {
	ac := &actorContext{}
	var err error
	if teamID != 0 {
		if ac.TeamRoles, err = s.repo.GetUserTeamRoles(ctx, actorID, teamID); err != nil {
			return nil, err
		}
	}
	if teamGameID != 0 {
		if ac.TeamGameRoles, err = s.repo.GetUserTeamGameRoles(ctx, actorID, teamGameID); err != nil {
			return nil, err
		}
	}
	if organizationID != 0 {
		if ac.OrganizationRoles, err = s.repo.GetUserOrganizationRoles(ctx, actorID, organizationID); err != nil {
			return nil, err
		}
	}
	if tournamentID != 0 {
		if ac.TournamentRoles, err = s.repo.GetUserTournamentRoles(ctx, actorID, tournamentID); err != nil {
			return nil, err
		}
	}
	return ac, nil
}

// tournamentAcceptsRegistrations validates the tournament state for new
// registrations: it must exist, not be finished/cancelled, and not have ended.
func tournamentAcceptsRegistrations(t *TournamentInfo) error {
	switch t.Status {
	case "DRAFT", "FINISHED", "CANCELLED":
		return rules.ErrTournamentNotOpen
	}
	if t.EndAt != nil {
		return rules.ErrTournamentNotOpen
	}
	return nil
}

// CreateRegistration registers a roster (team_game) to a tournament.
func (s *Service) CreateRegistration(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	tournamentID int64,
	teamGameID int64,
) (int64, error) {
	if identity.IsBanned(actorGlobalRoles) {
		return 0, errors.New("user is banned")
	}

	tournament, err := s.repo.GetTournamentInfo(ctx, tournamentID)
	if err != nil {
		return 0, err
	}
	if err := tournamentAcceptsRegistrations(tournament); err != nil {
		return 0, err
	}

	teamID, gameID, active, err := s.repo.GetTeamGameInfo(ctx, teamGameID)
	if err != nil {
		return 0, err
	}
	if !active {
		return 0, rules.ErrTeamGameInactive
	}
	if gameID != tournament.GameID {
		return 0, rules.ErrGameMismatch
	}

	teamRoles, err := s.repo.GetUserTeamRoles(ctx, actorID, teamID)
	if err != nil {
		return 0, err
	}
	if err := rules.CanRegisterTeamGame(teamRoles, actorGlobalRoles); err != nil {
		return 0, err
	}

	return s.repo.CreateRegistration(ctx, tournamentID, teamGameID, actorID)
}

// ListTournamentRegistrations lists the registrations of a tournament.
// Anonymous/non-privileged users only see ACCEPTED registrations.
func (s *Service) ListTournamentRegistrations(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	tournamentID int64,
) ([]RegistrationResponse, error) {
	tournament, err := s.repo.GetTournamentInfo(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	acceptedOnly := true
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		acceptedOnly = false
	} else {
		ac, err := s.buildActorContext(ctx, actorID, 0, 0, tournament.OrganizationID, tournament.ID)
		if err != nil {
			return nil, err
		}
		if err := rules.CanReviewRegistration(ac.OrganizationRoles, ac.TournamentRoles, actorGlobalRoles); err == nil {
			acceptedOnly = false
		}
	}

	return s.repo.ListRegistrationsByTournament(ctx, tournamentID, acceptedOnly)
}

// ReviewRegistration accepts or rejects a pending registration.
// Rejection deletes the registration; its roster rows cascade with it.
func (s *Service) ReviewRegistration(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	registrationID int64,
	decision string,
) error {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}
	if reg.Status != string(StatusPending) {
		return rules.ErrRegistrationNotPending
	}

	tournament, err := s.repo.GetTournamentInfo(ctx, reg.TournamentID)
	if err != nil {
		return err
	}
	ac, err := s.buildActorContext(ctx, actorID, 0, 0, tournament.OrganizationID, reg.TournamentID)
	if err != nil {
		return err
	}
	if err := rules.CanReviewRegistration(ac.OrganizationRoles, ac.TournamentRoles, actorGlobalRoles); err != nil {
		return err
	}

	switch decision {
	case "REJECT":
		// Hard delete: roster rows cascade, the roster's users are freed.
		return s.repo.DeleteRegistration(ctx, registrationID)
	case "ACCEPT":
		return s.repo.AcceptRegistration(ctx, registrationID, actorID)
	default:
		return rules.ErrRegistrationNotPending
	}
}

// CancelRegistration lets the registering team withdraw. The registration is
// deleted (roster cascades) so the roster can re-apply or register elsewhere.
func (s *Service) CancelRegistration(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	registrationID int64,
) error {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}

	teamID, _, _, err := s.repo.GetTeamGameInfo(ctx, reg.TeamGameID)
	if err != nil {
		return err
	}
	teamRoles, err := s.repo.GetUserTeamRoles(ctx, actorID, teamID)
	if err != nil {
		return err
	}
	if err := rules.CanCancelRegistration(actorID, reg.RegisteredBy, teamRoles, actorGlobalRoles); err != nil {
		return err
	}

	return s.repo.DeleteRegistration(ctx, registrationID)
}

// ListRegistrationRoster returns the players of a registration's roster.
func (s *Service) ListRegistrationRoster(ctx context.Context, registrationID int64) ([]RosterMemberResponse, error) {
	members, err := s.repo.ListRosterMembers(ctx, registrationID)
	if err != nil {
		return nil, err
	}
	out := make([]RosterMemberResponse, 0, len(members))
	for _, m := range members {
		resp := RosterMemberResponse{
			UserID:    m.UserID,
			Nickname:  m.Nickname,
			FirstName: m.FirstName,
			LastName:  m.LastName,
			IsStarter: m.IsStarter,
			AddedAt:   m.AddedAt,
		}
		if m.JerseyImage != nil {
			u := s.storage.URL(storage.FolderTeamGameJersey, *m.JerseyImage)
			resp.JerseyPhoto = &u
		}
		out = append(out, resp)
	}
	return out, nil
}

// AddRosterMember adds a player of the roster to the registration's roster.
// Enforces the "one roster per user per tournament" rule (the Donk rule).
func (s *Service) AddRosterMember(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	registrationID int64,
	userID int64,
	isStarter bool,
) error {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}

	teamID, _, _, err := s.repo.GetTeamGameInfo(ctx, reg.TeamGameID)
	if err != nil {
		return err
	}
	ac, err := s.buildActorContext(ctx, actorID, teamID, reg.TeamGameID, 0, 0)
	if err != nil {
		return err
	}
	if err := rules.CanManageRegistrationRoster(ac.TeamRoles, ac.TeamGameRoles, actorGlobalRoles); err != nil {
		return err
	}

	isMember, err := s.repo.IsTeamGameMember(ctx, userID, reg.TeamGameID)
	if err != nil {
		return err
	}
	if !isMember {
		return rules.ErrNotTeamGameMember
	}

	// Donk rule: a user can only be in one roster of the tournament, but the
	// roster he is already in may be this very registration (updating starter
	// state is done via remove+add).
	inTournament, err := s.repo.IsRosterMemberOfTournament(ctx, userID, reg.TournamentID)
	if err != nil {
		return err
	}
	if inTournament {
		return rules.ErrUserInAnotherRoster
	}

	return s.repo.AddRosterMember(ctx, registrationID, reg.TournamentID, userID, actorID, isStarter)
}

// RemoveRosterMember removes a player from the registration's roster.
func (s *Service) RemoveRosterMember(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	registrationID int64,
	userID int64,
) error {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}

	teamID, _, _, err := s.repo.GetTeamGameInfo(ctx, reg.TeamGameID)
	if err != nil {
		return err
	}
	ac, err := s.buildActorContext(ctx, actorID, teamID, reg.TeamGameID, 0, 0)
	if err != nil {
		return err
	}
	if err := rules.CanManageRegistrationRoster(ac.TeamRoles, ac.TeamGameRoles, actorGlobalRoles); err != nil {
		return err
	}

	return s.repo.RemoveRosterMember(ctx, registrationID, userID)
}

func ParseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
