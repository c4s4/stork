-- create database test

DROP DATABASE IF EXISTS test;
CREATE DATABASE test;
USE test;

-- create table user

DROP TABLE IF EXISTS user;
CREATE TABLE user (
	id INTEGER NOT NULL AUTO_INCREMENT,
	name VARCHAR(64) NOT NULL,
	PRIMARY KEY (id)
);
