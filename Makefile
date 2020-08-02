# Parent Makefiles https://github.com/c4s4/make

include ~/.make/Golang.mk
include .env
export

VERSION := "UNKNOWN"

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

test: go-build # Run test
	$(title)
	@$(BUILD_DIR)/stork -env=.env -init sql

docker: go-clean # Build docker image
	$(title)
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -f" -o $(BUILD_DIR)/stork .
	@docker build -t casa/stork:$(VERSION) .
	@docker tag casa/stork:$(VERSION) casa/stork:latest

test-docker: # Test docker image
	$(title)
	@docker run --network host --rm --volume=$(shell pwd)/sql:/sql --env-file=.env casa/stork -init /sql

publish: docker # Publish docker image
	$(title)
	@docker push casa/stork:$(VERSION)
	@docker push casa/stork:latest

release: go-tag publish go-deploy go-archive # Perform a release (must pass VERSION=X.Y.Z on command line)
	@echo "$(GRE)OK$(END) Release done!"
