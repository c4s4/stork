# Parent Makefiles https://github.com/c4s4/make

include ~/.make/Golang.mk
include .env
export

GONAME = "stork"

mysql: # Start Mysql
	$(title)
	@docker-compose up -d

shell: # Connect to Mysql
	$(title)
	@docker-compose exec mysql mysql -h$(MYSQL_HOSTNAME) -u$(MYSQL_USERNAME) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)

root: # Connect to Mysql as root
	$(title)
	@docker-compose exec mysql mysql -h$(MYSQL_HOSTNAME) -uroot -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE)

init: mysql # Initialize Mysql database
	$(title)
	@docker-compose exec mysql /bin/sh -c "mysql -h$(MYSQL_HOSTNAME) -uroot -p$(MYSQL_ROOT_PASSWORD) < /sql/init.sql"

test: go-build # Run test
	$(title)
	@$(BUILD_DIR)/stork -env=.env -init sql

docker: clean # Build docker image
	$(title)
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/stork .
	@docker build -t casa/stork .

test-docker: # Test docker image
	$(title)
	@docker run --network host --rm --volume=$(shell pwd)/sql:/sql --env-file=.env casa/stork -init /sql

publish: docker # Publish docker image
	$(title)
	@docker push casa/stork
