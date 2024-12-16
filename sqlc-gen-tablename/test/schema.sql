-- name: Authors
CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text NOT NULL,
  bio  text
);

-- name: EmployeesBla
CREATE TABLE employees (
  id       BIGSERIAL PRIMARY KEY,
  name     text NOT NULL,
  address  text
);
