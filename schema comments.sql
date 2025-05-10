-- SCHEMA: public

-- DROP SCHEMA IF EXISTS public ;

CREATE SCHEMA IF NOT EXISTS public
    AUTHORIZATION postgres;

DROP TABLE IF EXISTS comments;

CREATE TABLE IF NOT EXISTS comments
(
post_id BIGSERIAL PRIMARY KEY,
parent_id BIGINT NOT NULL,
contents TEXT NOT NULL,
creation_date TEXT NOT NULL,
URL TEXT NOT NULL
);

DELETE FROM comments;