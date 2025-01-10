-include .env
-include ../../../version
#####################################################################################
ADMIN_VERSION=$(admin_version)
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' 
.PHONY: help
#####################################################################################
## start all containers
start: 
	docker compose up -d dbmigration doorman-admin doorman test-api doorman-redis
	docker compose up -d doorman-admin --wait doorman-admin --timeout 10
.PHONY: start
## invoke integration tests
test/integration: start 
	docker compose up --build --exit-code-from integration-tests integration-tests
.PHONY: test/integration
## invoke unit tests
test/unit:  
	docker compose up --build --exit-code-from unit-tests unit-tests
.PHONY: test/unit
## clean everything, stops docker containers and removes them
clean:
	docker compose down
.PHONY: clean	
#####################################################################################

.EXPORT_ALL_VARIABLES:
