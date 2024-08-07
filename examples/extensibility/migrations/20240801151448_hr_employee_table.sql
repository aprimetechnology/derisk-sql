-- migrate:up
CREATE SCHEMA
    IF NOT EXISTS
       human_resources;
 GRANT USAGE
    ON SCHEMA human_resources
    TO PUBLIC;
 GRANT CREATE
    ON SCHEMA human_resources
    TO PUBLIC;
   SET SEARCH_PATH
    TO 'human_resources';

CREATE TABLE employee (
       id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
       first_name      varchar(100) NOT NULL,
       middle_name     varchar(100) NULL,
       last_name       varchar(100) NOT NULL,
       salary          numeric
);
INSERT INTO employee (first_name, last_name, salary)
VALUES ('John', 'Smith', 50000),
       ('Alice', 'Bob', 90000);
CREATE INDEX CONCURRENTLY first_name_idx
    ON employee(first_name);

-- migrate:down
  DROP INDEX CONCURRENTLY first_name_idx;
  DROP TABLE employee;
  DROP SCHEMA human_resources;
