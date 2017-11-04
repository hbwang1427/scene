CREATE TABLE sysinfo (
	dbver integer
);

CREATE TABLE "user" (
	id SERIAL PRIMARY KEY,
	name text NOT NULL UNIQUE,
	email text UNIQUE,
	password text,
	salt bigint,
	create_at timestamp
);

CREATE TABLE museums (
	id SERIAL PRIMARY KEY,
	name text NOT NULL UNIQUE,
	address text,
	"desc" text
);