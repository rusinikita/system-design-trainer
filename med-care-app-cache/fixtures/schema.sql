CREATE TABLE IF NOT EXISTS segment_types
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS articles
(
    id           SERIAL PRIMARY KEY,
    title        VARCHAR(255)             NOT NULL,
    content      TEXT                     NOT NULL,
    source       VARCHAR(255)             NOT NULL,
    type         VARCHAR(50)              NOT NULL,
    published_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- CREATE INDEX idx_articles_published_at ON articles(published_at DESC);

CREATE TABLE IF NOT EXISTS article_segments
(
    article_id      INTEGER NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    segment_id      INTEGER NOT NULL REFERENCES segment_types (id),
    relevance_score FLOAT   NOT NULL CHECK (relevance_score >= 0 AND relevance_score <= 1),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (article_id, segment_id)
);

CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_segments
(
    user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    segment_id    INTEGER NOT NULL REFERENCES segment_types (id),
    weight        FLOAT   NOT NULL CHECK (weight >= 0 AND weight <= 1),
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, segment_id)
);

CREATE TABLE IF NOT EXISTS care_plans
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER                  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       VARCHAR(255)             NOT NULL,
    description TEXT,
    status      VARCHAR(50)              NOT NULL CHECK (status IN ('active', 'completed', 'cancelled')),
    start_date  TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date    TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_dates CHECK (end_date > start_date)
);

-- CREATE INDEX idx_care_plans_user ON care_plans(user_id);
-- CREATE INDEX idx_care_plans_status ON care_plans(status) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS care_plan_steps
(
    id           SERIAL PRIMARY KEY,
    care_plan_id INTEGER                  NOT NULL REFERENCES care_plans (id) ON DELETE CASCADE,
    type         VARCHAR(50)              NOT NULL CHECK (type IN ('appointment', 'test', 'upload', 'checkup')),
    title        VARCHAR(255)             NOT NULL,
    description  TEXT,
    status       VARCHAR(50)              NOT NULL CHECK (status IN ('pending', 'completed', 'overdue')),
    due_date     TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata     JSONB                    DEFAULT '{}',
    order_number INTEGER                  NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_completion CHECK (
        (status = 'completed' AND completed_at IS NOT NULL) OR
        (status != 'completed' AND completed_at IS NULL)
        )
);

-- CREATE INDEX idx_care_plan_steps_plan ON care_plan_steps(care_plan_id);
-- CREATE INDEX idx_care_plan_steps_status ON care_plan_steps(status) WHERE status != 'completed';
-- CREATE INDEX idx_care_plan_steps_due_date ON care_plan_steps(due_date) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS read_articles
(
    user_id    INTEGER                  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    article_id INTEGER                  NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    read_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    is_saved   BOOLEAN                  NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE          DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, article_id)
);

-- CREATE INDEX idx_read_articles_user ON read_articles(user_id);
-- CREATE INDEX idx_read_articles_article ON read_articles(article_id);
-- CREATE INDEX idx_read_articles_read_at ON read_articles(read_at DESC);