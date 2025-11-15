
CREATE TABLE IF NOT EXISTS public.file_cache (
	id serial4 NOT NULL,
	source_url text NOT NULL,
	file_path text NOT NULL,
	content_type text NOT NULL,
	added_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT file_cache_pkey PRIMARY KEY (id),
	UNIQUE(source_url)
);

alter table public.file_cache add column if not exists file_size bigint null;
alter table public.file_cache add column if not exists last_read_at timestamptz default now();
