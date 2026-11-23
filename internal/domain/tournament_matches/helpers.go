package matches

import (
	"context"
	"encoding/json"
)

// jsonUnmarshal decodes JSON into v, reporting whether it succeeded.
func jsonUnmarshal(raw []byte, v interface{}) error {
	return json.Unmarshal(raw, v)
}

// actorRoles resolves the actor's organization and tournament roles for a
// tournament. Global staff callers skip the queries.
func (s *Service) actorRoles(ctx context.Context, actorID, tournamentID int64) (orgRoles, tournamentRoles []string, err error) {
	_, organizationID, err := s.repo.GetTournamentInfo(ctx, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	if orgRoles, err = s.repo.GetUserOrganizationRoles(ctx, actorID, organizationID); err != nil {
		return nil, nil, err
	}
	if tournamentRoles, err = s.repo.GetUserTournamentRoles(ctx, actorID, tournamentID); err != nil {
		return nil, nil, err
	}
	return orgRoles, tournamentRoles, nil
}

// toResponses enriches matches with team names for visualization.
func (s *Service) toResponses(ctx context.Context, ms []Match, tournamentID int64) ([]MatchResponse, error) {
	labels, err := s.repo.ListTeamLabels(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]TeamLabel, len(labels))
	for _, l := range labels {
		byID[l.RegistrationID] = l
	}
	team := func(id *int64) *TeamInfo {
		if id == nil {
			return nil
		}
		if l, ok := byID[*id]; ok {
			return &TeamInfo{RegistrationID: *id, TeamName: l.TeamName}
		}
		return &TeamInfo{RegistrationID: *id}
	}
	winner := func(id *int64) *TeamInfo {
		if id == nil {
			return nil
		}
		return team(id)
	}

	out := make([]MatchResponse, 0, len(ms))
	for _, m := range ms {
		resp := MatchResponse{
			ID:           m.ID,
			StageID:      m.StageID,
			TournamentID: m.TournamentID,
			RoundNumber:  m.RoundNumber,
			MatchNumber:  m.MatchNumber,
			TeamA:        team(m.TeamA),
			TeamB:        team(m.TeamB),
			BestOf:       m.BestOf,
			Status:       m.Status,
			ScoreA:       m.ScoreA,
			ScoreB:       m.ScoreB,
			Winner:       winner(m.Winner),
		}
		out = append(out, resp)
	}
	return out, nil
}
