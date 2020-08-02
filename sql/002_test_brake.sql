-- script to check that scripts run in a transaction

INSERT INTO test.user (name, email) VALUES ('test', 'test@example.com');

-- now an invalid query

FOO BAR;
