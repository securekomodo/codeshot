CREATE TABLE snippets (
  id          text PRIMARY KEY,
  owner_id    uuid NOT NULL REFERENCES users(id),
  language    text NOT NULL DEFAULT 'javascript',
  body        text NOT NULL,
  public      boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX snippets_owner_idx ON snippets (owner_id, created_at DESC);
