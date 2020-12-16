-- initialize test database

CREATE USER 'stork'@'localhost' IDENTIFIED BY 'stork';
GRANT ALL PRIVILEGES ON *.* TO 'stork'@'localhost';
FLUSH PRIVILEGES;
