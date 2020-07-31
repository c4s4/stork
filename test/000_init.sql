-- create table user

DROP DATABASE IF EXISTS test;
CREATE DATABASE test;

USE test;

CREATE TABLE user (
	id INTEGER NOT NULL AUTO_INCREMENT,
	name VARCHAR(64) NOT NULL,
	PRIMARY KEY (id)
);
