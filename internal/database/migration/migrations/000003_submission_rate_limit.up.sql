CREATE INDEX idx_submissions_user_task_submitted_at
    ON submissions (user_id, task_id, submitted_at DESC);
