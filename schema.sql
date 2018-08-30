<<<<<<< HEAD
CREATE USER aitour PASSWORD 'aitour' login;

CREATE DATABASE aitour OWNER aitour;

\connect aitour;

CREATE SEQUENCE global_id_sequence;
ALTER SEQUENCE global_id_sequence OWNER TO aitour;

CREATE OR REPLACE FUNCTION id_generator(OUT result bigint) AS $$
DECLARE
    our_epoch bigint := 1314220021721;
    seq_id bigint;
    now_millis bigint;
    -- the id of this DB shard, must be set for each
    -- schema shard you have - you could pass this as a parameter too
    shard_id int := 1;
BEGIN
    SELECT nextval('global_id_sequence') % 1024 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 23;
    result := result | (shard_id << 10);
    result := result | (seq_id);
END;
$$ LANGUAGE PLPGSQL;

ALTER FUNCTION id_generator OWNER TO aitour;

-- user --
CREATE TABLE IF NOT EXISTS ai_user (
	id BIGINT NOT NULL DEFAULT id_generator(),
	name TEXT,
	phone TEXT,
	email TEXT NOT NULL UNIQUE,
	password TEXT,
	create_at TIMESTAMP,
	update_at TIMESTAMP,
	enabled BOOLEAN,
	deleted BOOLEAN,
	activated BOOLEAN,

	CONSTRAINT uk_name UNIQUE(email),
	CONSTRAINT uk_id UNIQUE(id)
);
ALTER TABLE ai_user OWNER TO aitour;


-- user profile --
CREATE TABLE IF NOT EXISTS ai_user_profile (
	user_id BIGINT REFERENCES ai_user (id),
	nickname TEXT,
	avatar TEXT,
	lang TEXT
);
ALTER TABLE ai_user_profile OWNER TO aitour;


CREATE TABLE IF NOT EXISTS ai_user_oauth (
	user_id BIGINT REFERENCES ai_user (id),
	platform TEXT,
	openid TEXT,
	access_token TEXT,
	expires_at TIMESTAMP,

	CONSTRAINT uk_openid UNIQUE(user_id, platform, openid)
);
ALTER TABLE ai_user_profile OWNER TO aitour;

-- user album --
CREATE TABLE IF NOT EXISTS ai_album_photo (
	id BIGINT NOT NULL DEFAULT id_generator(),
	user_id BIGINT REFERENCES ai_user (id),
	url TEXT,
	width INTEGER,
	height INTEGER,
	memo TEXT,
	upload_at TIMESTAMP
);
=======
CREATE USER aitour PASSWORD 'aitour' login;

CREATE DATABASE aitour OWNER aitour;

\connect aitour;

CREATE SEQUENCE global_id_sequence;
ALTER SEQUENCE global_id_sequence OWNER TO aitour;

CREATE OR REPLACE FUNCTION id_generator(OUT result bigint) AS $$
DECLARE
    our_epoch bigint := 1314220021721;
    seq_id bigint;
    now_millis bigint;
    -- the id of this DB shard, must be set for each
    -- schema shard you have - you could pass this as a parameter too
    shard_id int := 1;
BEGIN
    SELECT nextval('global_id_sequence') % 1024 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 23;
    result := result | (shard_id << 10);
    result := result | (seq_id);
END;
$$ LANGUAGE PLPGSQL;

ALTER FUNCTION id_generator OWNER TO aitour;

-- user --
create table if not exists ai_user (
	id bigint not null default id_generator(),
	name text,
	phone text,
	email text not null unique,
	password text,
	create_at timestamp,
	update_at timestamp,
	enabled bool,
	deleted bool,
	activated bool,
	activate_key text,

	constraint uk_name unique(email)
);
ALTER TABLE ai_user OWNER TO aitour;

-- user album --
create table if not exists ai_album_photo (
	id bigint not null default id_generator(),
	user bigint,
	url text,
	memo text,
	upload_at timestamp
);
>>>>>>> 249b59b1b72034eb4adccebd65eb7e406909de5f
ALTER TABLE ai_album_photo OWNER TO aitour;