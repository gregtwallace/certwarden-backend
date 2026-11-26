-- CHANGES v12 to v13:
-- - certificates:
--    - Add `tech_phone` and `tech_email` fields/columns



-- +goose Up
ALTER TABLE certificates ADD COLUMN tech_phone text NOT NULL DEFAULT '';
ALTER TABLE certificates ADD COLUMN tech_email text NOT NULL DEFAULT '';



-- +goose Down
ALTER TABLE certificates DROP COLUMN tech_phone;
ALTER TABLE certificates DROP COLUMN tech_email;

