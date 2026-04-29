package postgres

const SchemaV1 = `
CREATE TABLE IF NOT EXISTS matches (
  match_id TEXT PRIMARY KEY,
  idempotency_key TEXT NOT NULL UNIQUE,
  payload_hash TEXT NOT NULL,
  schema_version TEXT NOT NULL,
  completion_status TEXT NOT NULL,
  completion_reason TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS match_events (
  match_id TEXT NOT NULL REFERENCES matches(match_id),
  event_index INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  event_timestamp TEXT,
  event_payload JSONB,
  PRIMARY KEY (match_id, event_index)
);

ALTER TABLE matches
  ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}'::jsonb;`

func Migrations() []string {
	return []string{SchemaV1}
}
