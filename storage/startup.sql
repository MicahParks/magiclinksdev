CREATE DATABASE magiclinksdev;
\connect magiclinksdev;


CREATE SCHEMA mld;
CREATE TABLE mld.setup
(
    id      BOOLEAN PRIMARY KEY               DEFAULT TRUE,
    setup   JSONB                    NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT startup_check_one_row CHECK (id)
);
INSERT INTO mld.setup (setup)
VALUES ('{
  "plaintextClaims": false,
  "plaintextJWK": false,
  "semver": "v0.2.0"
}');

CREATE TABLE mld.service_account
(
    id       BIGSERIAL PRIMARY KEY,
    uuid     UUID                     NOT NULL UNIQUE,
    api_key  UUID                     NOT NULL UNIQUE,
    aud      UUID                     NOT NULL UNIQUE,
    is_admin BOOLEAN                  NOT NULL DEFAULT FALSE,
    created  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX ON mld.service_account (uuid);
CREATE INDEX ON mld.service_account (api_key);
CREATE INDEX ON mld.service_account (aud);
CREATE INDEX ON mld.service_account (created);

CREATE TABLE mld.jwk
(
    id              BIGSERIAL PRIMARY KEY,
    assets          BYTEA                    NOT NULL,
    key_id          TEXT                     NOT NULL,
    signing_default BOOLEAN                  NOT NULL DEFAULT FALSE,
    created         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    alg             TEXT                     NOT NULL
);
CREATE INDEX ON mld.jwk (key_id);
CREATE INDEX ON mld.jwk (signing_default);
CREATE INDEX ON mld.jwk (alg);

CREATE TABLE mld.link
(
    id                 BIGSERIAL PRIMARY KEY,
    sa_id              BIGINT                   NOT NULL REFERENCES mld.service_account (id),
    expires            TIMESTAMP WITH TIME ZONE NOT NULL,
    jwt_claims         BYTEA                    NOT NULL,
    jwt_key_id         TEXT                     NOT NULL,
    jwt_signing_method TEXT                     NOT NULL,
    redirect_query_key TEXT                     NOT NULL,
    redirect_url       TEXT                     NOT NULL,
    secret             UUID                     NOT NULL UNIQUE,
    visited            TIMESTAMP WITH TIME ZONE,
    created            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX ON mld.link (expires);
CREATE INDEX ON mld.link (redirect_url);
CREATE INDEX ON mld.link (secret);
CREATE INDEX ON mld.link (sa_id);
CREATE INDEX ON mld.link (visited);
CREATE INDEX ON mld.link (created);

