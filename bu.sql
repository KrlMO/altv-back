-- public.al_country definition

-- Drop table

-- DROP TABLE al_country;

CREATE TABLE al_country (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(80) NOT NULL,
	iso_code varchar(2) NOT NULL,
	flag_path text NULL,
	CONSTRAINT al_country_pkey PRIMARY KEY (id)
);


-- public.al_game definition

-- Drop table

-- DROP TABLE al_game;

CREATE TABLE al_game (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(200) NOT NULL,
	description text NULL,
	logo_path text NULL,
	banner_path text NULL,
	year_founded int4 NULL,
	slug varchar(100) NOT NULL,
	active bool DEFAULT true NULL,
	color_hex varchar(7) NULL,
	popularity_rank int4 NULL,
	CONSTRAINT al_game_pkey PRIMARY KEY (id),
	CONSTRAINT al_game_slug_key UNIQUE (slug)
);


-- public.al_genre_e_sports definition

-- Drop table

-- DROP TABLE al_genre_e_sports;

CREATE TABLE al_genre_e_sports (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(100) NOT NULL,
	description text NULL,
	CONSTRAINT al_genre_e_sports_pkey PRIMARY KEY (id)
);


-- public.al_global_role definition

-- Drop table

-- DROP TABLE al_global_role;

CREATE TABLE al_global_role (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT al_global_role_pkey PRIMARY KEY (id)
);


-- public.al_organization_role definition

-- Drop table

-- DROP TABLE al_organization_role;

CREATE TABLE al_organization_role (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT al_organization_role_name_key UNIQUE (name),
	CONSTRAINT al_organization_role_pkey PRIMARY KEY (id)
);


-- public.al_platform definition

-- Drop table

-- DROP TABLE al_platform;

CREATE TABLE al_platform (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(100) NOT NULL,
	description text NULL,
	logo_path text NULL,
	active bool DEFAULT true NULL,
	CONSTRAINT al_platform_pkey PRIMARY KEY (id)
);


-- public.al_team_game_role definition

-- Drop table

-- DROP TABLE al_team_game_role;

CREATE TABLE al_team_game_role (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT al_team_game_role_name_key UNIQUE (name),
	CONSTRAINT al_team_game_role_pkey PRIMARY KEY (id)
);


-- public.al_team_role definition

-- Drop table

-- DROP TABLE al_team_role;

CREATE TABLE al_team_role (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT al_team_role_pkey PRIMARY KEY (id)
);


-- public.al_tournament_role definition

-- Drop table

-- DROP TABLE al_tournament_role;

CREATE TABLE al_tournament_role (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT al_tournament_role_name_key UNIQUE (name),
	CONSTRAINT al_tournament_role_pkey PRIMARY KEY (id)
);


-- public.al_user definition

-- Drop table

-- DROP TABLE al_user;

CREATE TABLE al_user (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	first_name varchar(50) NOT NULL,
	last_name varchar(50) NOT NULL,
	nickname varchar(50) NOT NULL,
	password_hash text NOT NULL,
	birth_date date NOT NULL,
	registered_at timestamp DEFAULT now() NULL,
	country_id int4 NOT NULL,
	email varchar(100) NOT NULL,
	profile_photo text NULL,
	banner_photo text NULL,
	active bool DEFAULT true NULL,
	updated_at timestamp DEFAULT now() NULL,
	CONSTRAINT al_user_email_key UNIQUE (email),
	CONSTRAINT al_user_nickname_key UNIQUE (nickname),
	CONSTRAINT al_user_pkey PRIMARY KEY (id),
	CONSTRAINT fk_user_country FOREIGN KEY (country_id) REFERENCES al_country(id)
);


-- public.al_user_global_role definition

-- Drop table

-- DROP TABLE al_user_global_role;

CREATE TABLE al_user_global_role (
	user_id int4 NOT NULL,
	role_id int4 NOT NULL,
	created_at timestamp DEFAULT now() NULL,
	created_by int4 NULL,
	CONSTRAINT al_user_global_role_pkey PRIMARY KEY (user_id, role_id),
	CONSTRAINT fk_ugr_created_by FOREIGN KEY (created_by) REFERENCES al_user(id),
	CONSTRAINT fk_ugr_role FOREIGN KEY (role_id) REFERENCES al_global_role(id),
	CONSTRAINT fk_ugr_user FOREIGN KEY (user_id) REFERENCES al_user(id)
);


-- public.al_organization definition

-- Drop table

-- DROP TABLE al_organization;

CREATE TABLE al_organization (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(100) NOT NULL,
	description text NULL,
	logo_path text NULL,
	banner_path text NULL,
	created_by int4 NOT NULL,
	active bool DEFAULT true NULL,
	created_at timestamp DEFAULT now() NOT NULL,
	updated_at timestamp DEFAULT now() NOT NULL,
	CONSTRAINT al_organization_pkey PRIMARY KEY (id),
	CONSTRAINT fk_org_owner FOREIGN KEY (created_by) REFERENCES al_user(id)
);


-- public.al_team definition

-- Drop table

-- DROP TABLE al_team;

CREATE TABLE al_team (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(100) NOT NULL,
	tag varchar(15) NOT NULL,
	description text NULL,
	logo_path text NULL,
	banner_path text NULL,
	country_id int4 NOT NULL,
	active bool DEFAULT true NULL,
	is_virtual bool DEFAULT false NULL,
	created_at timestamp DEFAULT now() NOT NULL,
	updated_at timestamp DEFAULT now() NOT NULL,
	created_by int4 NOT NULL,
	update_by int4 NOT NULL,
	CONSTRAINT al_team_name_key UNIQUE (name),
	CONSTRAINT al_team_pkey PRIMARY KEY (id),
	CONSTRAINT al_team_tag_key UNIQUE (tag),
	CONSTRAINT al_team_country_id_fkey FOREIGN KEY (country_id) REFERENCES al_country(id),
	CONSTRAINT al_team_created_by_fkey FOREIGN KEY (created_by) REFERENCES al_user(id),
	CONSTRAINT al_team_update_by_fkey FOREIGN KEY (update_by) REFERENCES al_user(id)
);


-- public.al_team_game definition

-- Drop table

-- DROP TABLE al_team_game;

CREATE TABLE al_team_game (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	team_id int4 NOT NULL,
	game_id int4 NOT NULL,
	"name" varchar(150) NOT NULL,
	logo_path text NULL,
	banner_path text NULL,
	active bool DEFAULT true NULL,
	created_at timestamp DEFAULT now() NOT NULL,
	updated_at timestamp DEFAULT now() NOT NULL,
	created_by int4 NOT NULL,
	update_by int4 NOT NULL,
	CONSTRAINT al_team_game_pkey PRIMARY KEY (id),
	CONSTRAINT al_team_game_team_id_game_id_key UNIQUE (team_id, game_id),
	CONSTRAINT al_team_game_game_id_fkey FOREIGN KEY (game_id) REFERENCES al_game(id),
	CONSTRAINT al_team_game_team_id_fkey FOREIGN KEY (team_id) REFERENCES al_team(id)
);


-- public.al_tournament definition

-- Drop table

-- DROP TABLE al_tournament;

CREATE TABLE al_tournament (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	organization_id int4 NOT NULL,
	game_id int4 NOT NULL,
	platform_id int4 NULL,
	genre_id int4 NULL,
	"name" text NOT NULL,
	description text NULL,
	start_at timestamp NOT NULL,
	end_at timestamp NULL,
	status varchar(20) DEFAULT 'DRAFT'::character varying NULL,
	created_at timestamp DEFAULT now() NOT NULL,
	updated_at timestamp DEFAULT now() NOT NULL,
	created_by int4 NOT NULL,
	color_hex varchar(7) NULL,
	banner_path text NULL,
	logo_path text NULL,
	CONSTRAINT al_tournament_pkey PRIMARY KEY (id),
	CONSTRAINT al_tournament_created_by_fkey FOREIGN KEY (created_by) REFERENCES al_user(id),
	CONSTRAINT al_tournament_game_id_fkey FOREIGN KEY (game_id) REFERENCES al_game(id),
	CONSTRAINT al_tournament_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES al_genre_e_sports(id),
	CONSTRAINT al_tournament_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES al_organization(id),
	CONSTRAINT al_tournament_platform_id_fkey FOREIGN KEY (platform_id) REFERENCES al_platform(id)
);


-- public.al_tournament_roster definition

-- Drop table

-- DROP TABLE al_tournament_roster;

CREATE TABLE al_tournament_roster (
	user_id int4 NOT NULL,
	tournament_id int4 NOT NULL,
	team_game_id int4 NOT NULL,
	is_starter bool DEFAULT true NOT NULL,
	added_at timestamp DEFAULT now() NOT NULL,
	added_by int4 NOT NULL,
	CONSTRAINT al_tournament_roster_pkey PRIMARY KEY (user_id, tournament_id),
	CONSTRAINT al_tournament_roster_added_by_fkey FOREIGN KEY (added_by) REFERENCES al_user(id),
	CONSTRAINT al_tournament_roster_team_game_id_fkey FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
	CONSTRAINT al_tournament_roster_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
	CONSTRAINT al_tournament_roster_user_id_fkey FOREIGN KEY (user_id) REFERENCES al_user(id)
);


-- public.al_user_organization_role definition

-- Drop table

-- DROP TABLE al_user_organization_role;

CREATE TABLE al_user_organization_role (
	user_id int4 NOT NULL,
	organization_id int4 NOT NULL,
	role_id int4 NOT NULL,
	assigned_at timestamp DEFAULT now() NOT NULL,
	assigned_by int4 NOT NULL,
	active bool DEFAULT true NULL,
	CONSTRAINT al_user_organization_role_pkey PRIMARY KEY (user_id, organization_id, role_id),
	CONSTRAINT al_user_organization_role_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES al_user(id),
	CONSTRAINT al_user_organization_role_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES al_organization(id),
	CONSTRAINT al_user_organization_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES al_organization_role(id),
	CONSTRAINT al_user_organization_role_user_id_fkey FOREIGN KEY (user_id) REFERENCES al_user(id)
);


-- public.al_user_team_game_role definition

-- Drop table

-- DROP TABLE al_user_team_game_role;

CREATE TABLE al_user_team_game_role (
	user_id int4 NOT NULL,
	team_game_id int4 NOT NULL,
	role_id int4 NOT NULL,
	assigned_at timestamp DEFAULT now() NOT NULL,
	assigned_by int4 NOT NULL,
	CONSTRAINT al_user_team_game_role_pkey PRIMARY KEY (user_id, team_game_id, role_id),
	CONSTRAINT al_user_team_game_role_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES al_user(id),
	CONSTRAINT al_user_team_game_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES al_team_game_role(id),
	CONSTRAINT al_user_team_game_role_team_game_id_fkey FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
	CONSTRAINT al_user_team_game_role_user_id_fkey FOREIGN KEY (user_id) REFERENCES al_user(id)
);


-- public.al_user_tournament_role definition

-- Drop table

-- DROP TABLE al_user_tournament_role;

CREATE TABLE al_user_tournament_role (
	user_id int4 NOT NULL,
	tournament_id int4 NOT NULL,
	role_id int4 NOT NULL,
	assigned_at timestamp DEFAULT now() NOT NULL,
	active bool DEFAULT true NULL,
	assigned_by int4 NOT NULL,
	CONSTRAINT al_user_tournament_role_pkey PRIMARY KEY (user_id, tournament_id, role_id),
	CONSTRAINT al_user_tournament_role_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES al_user(id),
	CONSTRAINT al_user_tournament_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES al_tournament_role(id),
	CONSTRAINT al_user_tournament_role_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
	CONSTRAINT al_user_tournament_role_user_id_fkey FOREIGN KEY (user_id) REFERENCES al_user(id)
);