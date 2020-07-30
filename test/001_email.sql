-- add email column

ALTER TABLE user
 ADD COLUMN email VARCHAR(100) NOT NULL;
