-- SCHEMA: public

-- DROP SCHEMA IF EXISTS public ;

CREATE SCHEMA IF NOT EXISTS public
    AUTHORIZATION postgres;

DROP TABLE IF EXISTS public.news;

CREATE TABLE IF NOT EXISTS public.news
(
id BIGSERIAL PRIMARY KEY,
title TEXT NOT NULL,
contents TEXT NOT NULL,
publishing_date TEXT NOT NULL,
URL TEXT NOT NULL
);

DELETE FROM public.news;