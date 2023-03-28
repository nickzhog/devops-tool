CREATE TABLE IF NOT EXISTS public.metrics (
    id TEXT NOT NULL, 
    type TEXT NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT,
    PRIMARY KEY (id, type)
);