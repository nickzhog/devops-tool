CREATE TABLE IF NOT EXISTS public.metrics (
    id text not null, 
    type text not null,
    value double precision,
    delta BIGINT,
    PRIMARY KEY (id, type)
);