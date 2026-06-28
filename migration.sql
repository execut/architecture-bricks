CREATE TABLE product (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    approve_reason TEXT,
    rejection_reason TEXT,
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE event (
    id UUID PRIMARY KEY,
    entry_id UUID NOT NULL,
    event TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX event_entry_id_idx ON event (entry_id);
