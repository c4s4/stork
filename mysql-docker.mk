# MySQL with Docker

mysql: # Start Mysql
	$(title)
	@docker-compose up -d

stop: # Stop Mysql
	$(title)
	@docker-compose down

shell: # Connect to Mysql
	$(title)
	@docker-compose exec mysql mysql -h$(MYSQL_HOSTNAME) -u$(MYSQL_USERNAME) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)

root: # Connect to Mysql as root
	$(title)
	@docker-compose exec mysql mysql -h$(MYSQL_HOSTNAME) -uroot -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE)

init: mysql # Initialize Mysql database
	$(title)
	@docker-compose exec mysql /bin/sh -c "mysql -h$(MYSQL_HOSTNAME) -uroot -p$(MYSQL_ROOT_PASSWORD) < /sql/init.sql"

test-docker: go-docker # Test docker image
	$(title)
	@docker run --network host --rm --volume=$(shell pwd)/sql:/sql --env-file=.env casa/stork -init /sql
