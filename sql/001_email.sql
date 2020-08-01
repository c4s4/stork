-- add email column

ALTER TABLE test.user
 ADD COLUMN email VARCHAR(100) NOT NULL;
