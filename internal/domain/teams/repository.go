package teams

import (
	"arena/internal/domain/teams/rules"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CountryExists(ctx context.Context, countryID int32) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		select exists(
			select 1 from al_country where id = $1
		)
	`, countryID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) CreateTeam(ctx context.Context, t Team) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO al_team
		(name, tag, description, logo_path, banner_path, country_id, is_virtual, active, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()) returning id
	`, t.Name, t.Tag, t.Description, t.LogoPath, t.BannerPath, t.CountryId, t.IsVirtual, t.Active, t.CreatedBy, t.UpdatedBy).Scan(&t.ID)
	if err != nil {
		return err
	}
	const teamRole = "TEAM_OWNER"
	err = r.InsertUserTeamRole(ctx, tx, t.CreatedBy, t.ID, t.CreatedBy, teamRole)

	return tx.Commit(ctx)
}

func (r *Repository) InsertUserTeamRole(
	ctx context.Context,
	tx pgx.Tx,
	userID, teamID, assignedBy int64,
	roleCode string,
) error {
	idRole, err := r.GetIdTeamRole(ctx, tx, roleCode)

	if err != nil || idRole == -1 {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_team_role
		(user_id, team_id, role_id, assigned_at, assigned_by)
		values
		($1, $2, $3, NOW(), $4)
	`, userID, teamID, idRole, assignedBy)

	if err != nil {
		return err
	}

	err = r.InsertUserTeamRoleHistory(ctx, tx, userID, teamID, assignedBy, idRole, "ASSIGNED")

	return err
}

func (r *Repository) GetIdTeamRole(ctx context.Context, tx pgx.Tx, roleName string) (int32, error) {
	var idRole int32
	err := tx.QueryRow(ctx, `
		SELECT id FROM al_team_role WHERE name = $1
	`, roleName).Scan(&idRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, errors.New("role does not exist")
		}
		return -1, err
	}
	return idRole, nil
}

func (r *Repository) GetTeamByID(ctx context.Context, teamID string) (*TeamReadModel, error) {
	var t TeamReadModel

	err := r.db.QueryRow(ctx, `
		SELECT
			at.id, at.name, at.tag, at.description, at.logo_path, at.banner_path,
			at.country_id, at.is_virtual, at.active,
			ac.id, ac.name, ac.iso_code, ac.flag_path
		FROM al_team at
		JOIN al_country ac ON at.country_id = ac.id
		WHERE at.id = $1
	`, teamID).Scan(
		&t.Team.ID,
		&t.Team.Name,
		&t.Team.Tag,
		&t.Team.Description,
		&t.Team.LogoPath,
		&t.Team.BannerPath,
		&t.Team.CountryId,
		&t.Team.IsVirtual,
		&t.Team.Active,
		&t.Country.ID,
		&t.Country.Name,
		&t.Country.IsoCode,
		&t.Country.FlagPath,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetTeams(ctx context.Context, filter TeamFilter) ([]TeamReadModel, int64, error) {
	query := `
		SELECT
			t.id,
			t.name,
			t.tag,
			t.description,
			t.logo_path,
			t.country_id,
			t.is_virtual,
			t.active,

			ac.id        AS country_id,
			ac.name      AS country_name,
			ac.iso_code,
			ac.flag_path

		FROM al_team t
		JOIN al_country ac 
			ON ac.id = t.country_id

		JOIN al_user_team_role utr
			ON utr.team_id = t.id

		JOIN al_team_role tr
			ON tr.id = utr.role_id
			AND tr.name = 'TEAM_OWNER'

		WHERE
			($1::INT IS NULL OR t.country_id = $1::INT)
		AND ($2::BIGINT IS NULL OR utr.user_id = $2::BIGINT)
		AND ($3::BOOLEAN IS NULL OR t.active = $3::BOOLEAN)
		AND ($4::TEXT IS NULL OR t.name ILIKE '%' || $4 || '%')
		AND ($5::TEXT IS NULL OR t.tag  ILIKE '%' || $5 || '%')

		ORDER BY t.id
		LIMIT $6 OFFSET $7;
	`

	rows, err := r.db.Query(ctx, query,
		filter.CountryID,
		filter.OwnerID,
		filter.Active,
		filter.NameLike,
		filter.TagLike,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var teams []TeamReadModel
	for rows.Next() {
		var t TeamReadModel
		err := rows.Scan(
			&t.Team.ID,
			&t.Team.Name,
			&t.Team.Tag,
			&t.Team.Description,
			&t.Team.LogoPath,
			&t.Team.CountryId,
			&t.Team.IsVirtual,
			&t.Team.Active,

			&t.Country.ID,
			&t.Country.Name,
			&t.Country.IsoCode,
			&t.Country.FlagPath,
		)
		if err != nil {
			return nil, 0, err
		}
		teams = append(teams, t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	total := int64(len(teams))

	return teams, total, nil
}

func (r *Repository) GetUserTeamRoles(ctx context.Context, userID, teamID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		SELECT tr.name
		FROM al_user_team_role utr
		JOIN al_team_role tr ON utr.role_id = tr.id
		WHERE utr.user_id = $1 AND utr.team_id = $2
	`, userID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, nil
}

func (r *Repository) UpdateTeam(ctx context.Context, t Team) error {
	set := []string{}
	args := []interface{}{}
	i := 1

	if t.Name != "" {
		set = append(set, "name = $"+strconv.Itoa(i))
		args = append(args, t.Name)
		i++
	}
	if t.Tag != "" {
		set = append(set, "tag = $"+strconv.Itoa(i))
		args = append(args, t.Tag)
		i++
	}
	if t.Description != "" {
		set = append(set, "description = $"+strconv.Itoa(i))
		args = append(args, t.Description)
		i++
	}
	if t.LogoPath != "" {
		set = append(set, "logo_path = $"+strconv.Itoa(i))
		args = append(args, t.LogoPath)
		i++
	}
	if t.BannerPath != "" {
		set = append(set, "banner_path = $"+strconv.Itoa(i))
		args = append(args, t.BannerPath)
		i++
	}
	if t.CountryId != 0 {
		set = append(set, "country_id = $"+strconv.Itoa(i))
		args = append(args, t.CountryId)
		i++
	}

	set = append(set, "is_virtual = $"+strconv.Itoa(i))
	args = append(args, t.IsVirtual)
	i++

	set = append(set, "updated_by = $"+strconv.Itoa(i))
	args = append(args, t.UpdatedBy)
	i++

	if len(set) == 0 {
		return errors.New("nothing to update")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := "UPDATE al_team SET " + strings.Join(set, ", ") + ", updated_at = NOW() WHERE id = $" + strconv.Itoa(i)
	args = append(args, t.ID)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) AddMemberToTeam(
	ctx context.Context,
	userID, teamID, assignedBy int64,
	roleCode string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = r.InsertUserTeamRole(ctx, tx, userID, teamID, assignedBy, roleCode)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) TeamRoleExists(ctx context.Context, tx pgx.Tx, roleName string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM al_team_role WHERE name = $1
		)
	`, roleName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) UpdateUserTeamRole(
	ctx context.Context,
	userId, teamId, assigned_by int64,
	roleName string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	idTeamRole, err := r.GetIdTeamRole(ctx, tx, roleName)

	if err != nil {
		return err
	}

	if err := r.ClosePreviousPeriodUserTeamRole(ctx, tx, userId, teamId); err != nil {
		return err
	}

	query := `
		update al_user_team_role
		set 
			role_id = $1,
			assigned_at = NOW(),
			assigned_by = $2
		where user_id = $3 and team_id = $4
	`
	_, err = tx.Exec(ctx, query, idTeamRole, assigned_by, userId, teamId)

	if err != nil {
		return err
	}

	err = r.InsertUserTeamRoleHistory(ctx, tx, userId, teamId, assigned_by, idTeamRole, "UPDATED")

	return tx.Commit(ctx)
}

func (r *Repository) ClosePreviousPeriodUserTeamRole(
	ctx context.Context,
	tx pgx.Tx,
	userId, teamId int64) error {
	_, err := tx.Exec(ctx, `
		UPDATE al_user_team_role_history SET assigned_to = NOW()
		WHERE user_id = $1 AND team_id = $2 AND assigned_to is NULL
	`, userId, teamId)
	return err
}

func (r *Repository) InsertUserTeamRoleHistory(
	ctx context.Context,
	tx pgx.Tx,
	userId, teamId, assignedBy int64,
	idRole int32,
	action string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO al_user_team_role_history
		(user_id, team_id, role_id, action, assigned_from, assigned_to, assigned_by)
		values
		($1, $2, $3, $4, NOW(), NULL, $5)
	`, userId, teamId, idRole, action, assignedBy)
	return err
}

func (r *Repository) DeleteUserFromTeam(
	ctx context.Context,
	userId, teamId, assigned_by int64,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := r.CloseUserTeamRoleAsRemoved(ctx, tx, userId, teamId, assigned_by); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM al_user_team_role
		WHERE user_id = $1 AND team_id = $2
	`, userId, teamId)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetUserTeamRole(
	ctx context.Context,
	userId, teamId int64,
) (string, error) {
	var actualRole string
	err := r.db.QueryRow(ctx, `
		SELECT atr.name FROM al_user_team_role autr
		JOIN al_team_role atr ON autr.role_id = atr.id
		WHERE autr.user_id = $1 AND autr.team_id = $2
	`, userId, teamId).Scan(&actualRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", rules.ErrUserNotMember
		}
		return "", err
	}

	return actualRole, err
}

func (r *Repository) CloseUserTeamRoleAsRemoved(
	ctx context.Context,
	tx pgx.Tx,
	userId, teamId, assignedBy int64,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE al_user_team_role_history
		SET
			assigned_to = NOW(),
			action = 'REMOVED',
			assigned_by = $3
		WHERE user_id = $1
		  AND team_id = $2
		  AND assigned_to IS NULL
	`, userId, teamId, assignedBy)
	return err
}

func (r *Repository) GetGameBySlug(
	ctx context.Context,
	gameSlug string,
) (int32, error) {
	var gameId int32
	err := r.db.QueryRow(ctx, `
			SELECT ag.id FROM al_game ag
			WHERE ag.slug = $1
		`, gameSlug).Scan(&gameId)
	if err != nil {
		return -1, err
	}
	return gameId, nil
}

func (r *Repository) CreateTeamGame(
	ctx context.Context,
	teamId, createdBy int64,
	gameId int32,
	nameTeamGame, logoPath, bannerPath string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	query := `
		INSERT INTO al_team_game(team_id, game_id, name, logo_path, banner_path, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := tx.Exec(ctx, query, teamId, gameId, nameTeamGame, logoPath, bannerPath, createdBy, createdBy); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
