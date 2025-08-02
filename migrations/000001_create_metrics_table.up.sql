
CREATE TABLE IF NOT EXISTS metrics
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ), 
    name text NOT NULL,
    delta bigint,
    value double precision,
    mtype text NOT NULL,
    CONSTRAINT metrics_pkey PRIMARY KEY (name, mtype)
);


CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);

CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(mtype);