CREATE UNIQUE INDEX ux_active_user_team_role
ON al_user_team_role_history (user_id, team_id)
WHERE assigned_to IS NULL;
