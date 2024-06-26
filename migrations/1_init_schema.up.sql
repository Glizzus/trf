CREATE TYPE rating AS ENUM (
    'True', 'Mostly True', 'Mostly False',
    'False', 'Legit', 'Fake', 'Correct Attribution',
    'Misattributed', 'Unproven', 'Unfounded',
    'Outdated', 'Miscaptioned', 'Legend',
    'Scam', 'Labeled Satire', 'Originated as Satire',
    'Research in Progress', 'Mixture',
    'Lost Legend', 'Recall'
);

COMMENT ON TYPE rating IS 'The rating of an article.
This is what Snopes uses to rate the truthfulness of articles.
A list can be found here: https://www.snopes.com/fact-check-ratings/';

CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE CONSTRAINT slug_not_empty CHECK (slug <> ''),
    title TEXT NOT NULL,
    subtitle TEXT NOT NULL,
    date DATE NOT NULL,

    question TEXT NOT NULL,
    rating RATING NOT NULL,
    context TEXT,

    content TEXT[] NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON CONSTRAINT slug_not_empty ON articles IS 'The slug of a Snopes article cannot be empty';

COMMENT ON TABLE articles IS 'Articles that Snopes has written and rated';
COMMENT ON COLUMN articles.slug IS 'The slug of the article.
It is used as the primary identifier of an article.
For example, the slug of https://www.snopes.com/fact-check/biden-banned-tiktok-in-us/ is "biden-banned-tiktok-in-us"';

-- We make this a new table instead of adding a column to the articles table
-- because there are semantic differences in the application. In normal use,
-- there is no reason to query both articles and spoofs at the same time.
CREATE TABLE spoofs (
    id SERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE REFERENCES articles (slug),
    rating RATING NOT NULL,
    content TEXT NOT NULL,

    templated BOOLEAN NOT NULL DEFAULT FALSE
);

COMMENT ON TABLE spoofs IS 'Articles that the application
has written to spoof an article from Snopes';

COMMENT ON COLUMN spoofs.slug IS 'The slug of the article
that this spoof is based on';
