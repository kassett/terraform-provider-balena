.PHONY: build
build:
	go build

.PHONY: init-integration-tests
init-integration-tests: build
	@echo
