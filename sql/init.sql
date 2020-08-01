-- initialize test database

CREATE USER 'stork'@'%' IDENTIFIED BY 'stork';
GRANT ALL PRIVILEGES ON *.* TO 'stork'@'%';
FLUSH PRIVILEGES;
