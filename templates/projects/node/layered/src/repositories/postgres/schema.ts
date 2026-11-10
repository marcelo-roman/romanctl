export const schema = `
CREATE TABLE IF NOT EXISTS tasks (
    id         uuid PRIMARY KEY,
    title      text        NOT NULL,
    done       boolean     NOT NULL DEFAULT false,
    due_date   date,
    created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS tasks_created_at ON tasks (created_at DESC);

-- romanctl:tables
`;
