# Parent Makefiles https://github.com/c4s4/make

include ~/.make/Golang.mk
include .env
export

GONAME = "stork"

mysql: # Start Mysql
	$(title)
	@cd test && docker-compose up -d

root: # Connect to Mysql as root
	$(title)
	@cd test && docker-compose exec mysql /bin/sh -c "mysql -h$(MYSQL_HOSTNAME) -uroot -p$(MYSQL_ROOT_PASSWORD)";

init: mysql # Initialize Mysql database
	$(title)
	@cd test && docker-compose exec mysql /bin/sh -c "mysql -h$(MYSQL_HOSTNAME) -uroot -p$(MYSQL_ROOT_PASSWORD) < /sql/init.sql";

test: go-build # Run test
	$(title)
	@$(BUILD_DIR)/stork -env=.env -init test

shell: # Open a mysql shell
	$(title)
	@cd test && docker-compose exec mysql mysql -h$(MYSQL_HOSTNAME) -u$(MYSQL_USERNAME) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)

docker: clean # Build docker image
	$(title)
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/stork .
	@docker build -t casa/stork .

publish: docker # Publish docker image
	$(title)
	@docker push casa/stork
