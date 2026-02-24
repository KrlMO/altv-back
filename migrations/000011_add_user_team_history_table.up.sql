-- al_user_team_role_history

CREATE TABLE al_user_team_role_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    role_id INT NOT NULL,
    action VARCHAR(30) NOT NULL,
    assigned_from TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_to TIMESTAMP,
    assigned_by BIGINT NOT NULL,

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (team_id) REFERENCES al_team(id),
    FOREIGN KEY (role_id) REFERENCES al_team_role(id),
    FOREIGN KEY (assigned_by) REFERENCES al_user(id)
);