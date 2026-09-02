-- CHANGES v5 to v6:
-- - certificates:
--     - If a certificate with the name `legocerthub` exists, rename it to `serverdefault`
-- Note: DB Schema doesn't actually change from 5 to 6.


-- +goose Up


-- rename cert
UPDATE certificates
  SET name = 'serverdefault'
  WHERE name = 'legocerthub';


-- +goose Down


-- rename cert
UPDATE certificates
  SET name = 'legocerthub'
  WHERE name = 'serverdefault';
