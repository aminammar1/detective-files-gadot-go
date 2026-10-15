CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email citext UNIQUE NOT NULL,
  username text UNIQUE NOT NULL,
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE player_profiles (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  display_name text NOT NULL,
  xp int NOT NULL DEFAULT 0,
  level int NOT NULL DEFAULT 1,
  rank text NOT NULL DEFAULT 'Rookie',
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE investigations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  case_id text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  started_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz,
  UNIQUE(user_id, case_id)
);

CREATE TABLE discovered_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investigation_id uuid NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
  evidence_id text NOT NULL,
  discovered_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(investigation_id, evidence_id)
);

CREATE TABLE interrogation_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investigation_id uuid NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
  suspect_id text NOT NULL,
  topic_id text,
  evidence_id text,
  kind text NOT NULL,
  response_text text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE investigation_actions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investigation_id uuid NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
  action text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE case_results (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  case_id text NOT NULL,
  suspect_id text NOT NULL,
  correct bool NOT NULL,
  proven bool NOT NULL,
  score int NOT NULL,
  xp_awarded int NOT NULL,
  completed_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(user_id, case_id)
);

CREATE TABLE achievements (
  id text PRIMARY KEY,
  name text NOT NULL,
  description text NOT NULL
);

CREATE TABLE player_achievements (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  achievement_id text NOT NULL REFERENCES achievements(id),
  unlocked_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(user_id, achievement_id)
);

CREATE INDEX idx_investigations_user ON investigations(user_id);
CREATE INDEX idx_discovered_evidence_inv ON discovered_evidence(investigation_id);
CREATE INDEX idx_interrogation_events_inv ON interrogation_events(investigation_id);
