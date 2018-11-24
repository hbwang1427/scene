CREATE USER aitour PASSWORD 'aitour' login;

CREATE DATABASE aitour_test OWNER aitour;

\connect aitour_test;

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
	feature bytea,
	upload_at TIMESTAMP
);
ALTER TABLE ai_album_photo OWNER TO aitour;


-- user uploaded porclain photos --
CREATE TABLE IF NOT EXISTS ai_porcelain_photo (
	id BIGINT NOT NULL PRIMARY KEY,
	url TEXT,
	class SMALLINT,
	upload_at TIMESTAMP,
	upload_from TEXT,
	file_hash TEXT UNIQUE
);
ALTER TABLE ai_porcelain_photo OWNER TO aitour;


--ai_museum--
CREATE TABLE IF NOT EXISTS ai_museum
(
    id bigint PRIMARY KEY DEFAULT id_generator(),
    name text,
    address text,
    city text,
    country text,
    lat numeric,
    lng numeric
);
ALTER TABLE ai_museum OWNER TO aitour;


-- ai_artist--
CREATE TABLE  IF NOT EXISTS ai_artist
(
    id bigint PRIMARY KEY DEFAULT id_generator(),
    first_name text,
    last_name text,
    middle_name text,
    country text,
    birth_year date,
    death_year date,
    gender text,
    bio text
);
ALTER TABLE ai_artist OWNER TO aitour;


--ai_art--
CREATE TABLE IF NOT EXISTS ai_art
(
    id bigint PRIMARY KEY DEFAULT id_generator(),
    museum_id bigint REFERENCES ai_museum(id),
    artist_id bigint,
    creation_year date,
    title text,
    category text,
    price numeric
);
ALTER TABLE ai_art OWNER TO aitour;


--ai_art_photo--
CREATE TABLE IF NOT EXISTS ai_art_photo (
	id bigint PRIMARY KEY DEFAULT id_generator(),
	art_id bigint REFERENCES ai_art(id),
	url text,
	feature bytea,
	width int,
	height int
);
ALTER TABLE ai_art_photo OWNER TO aitour;


--ai_art_media--
CREATE TABLE IF NOT EXISTS ai_art_media (
	id bigint PRIMARY KEY DEFAULT id_generator(),
	art_id bigint REFERENCES ai_art(id),
	type int,
	url text,
	size int,
	duration int
);
ALTER TABLE ai_art_media OWNER TO aitour;


--ai_art_memo--
CREATE TABLE IF NOT EXISTS ai_art_memo (
	art_id bigint REFERENCES ai_art(id),
	lang varchar(10),
	memo text,
	CONSTRAINT ai_art_memo_pk PRIMARY KEY(art_id),
	CONSTRAINT ai_art_memo_uk UNIQUE (art_id, lang)
);
ALTER TABLE ai_art_memo OWNER TO aitour;
