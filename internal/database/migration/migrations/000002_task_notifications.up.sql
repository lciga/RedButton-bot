CREATE TABLE task_notifications (
    user_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON UPDATE CASCADE ON DELETE CASCADE,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, task_id)
);

CREATE INDEX idx_task_notifications_task_id ON task_notifications (task_id);
