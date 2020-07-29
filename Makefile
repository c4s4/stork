# Parent Makefiles https://github.com/c4s4/make

include ~/.make/Golang.mk
include .env
export

GONAME = "stork"

mysql: # Start Mysql
	$(title)
	@cd test && docker-compose up -d

test: go-build # Run test
	$(title)
	@$(BUILD_DIR)/stork -env=.env test

shell: # Open a mysql shell
	$(title)
	@cd test && docker-compose exec mysql mysql -h$(MYSQL_HOSTNAME) -u$(MYSQL_USERNAME) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)
