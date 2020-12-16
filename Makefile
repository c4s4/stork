# Parent Makefiles https://github.com/c4s4/make

include ~/.make/Golang.mk
include config.mk
include .env
export

VERSION := "UNKNOWN"

test: go-build # Run test
	$(title)
	@$(BUILD_DIR)/stork -env=.env -init sql

version: # Check that version was passed on command line
	$(title)
	@if [ "$(VERSION)" = "UNKNOWN" ]; then \
		echo "$(RED)ERROR$(END) you must pass VERSION=X.Y.Z on command line to release"; \
		exit 1; \
	fi

release: version go-tag go-publish go-deploy go-archive # Perform a release (must pass VERSION=X.Y.Z on command line)
	@echo "$(GRE)OK$(END) Release done!"
