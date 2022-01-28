-include .env
-include ../../../.version
#####################################################################################
ADMIN_VERSION=$(admin_version).$(shell git rev-list --count HEAD)
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
	docker-compose up -d mongo
	$(MAKE) mongo/prepare 
	docker-compose up -d doorman-admin
.PHONY: start
## invoke integration tests
test/integration: start 
	docker-compose up --build --exit-code-from integration-tests integration-tests
.PHONY: test/integration
## invoke unit tests
test/unit:  
	docker-compose up --build --exit-code-from unit-tests unit-tests
.PHONY: test/unit
## clean everything, stops docker containers and removes them
clean:
	docker-compose stop && docker-compose rm -f
.PHONY: clean	
#####################################################################################
mongo/prepare:
	until docker-compose exec -T mongo /usr/bin/mongo -u $(MONGO_USER) -p $(MONGO_PASS) --eval "rs.initiate()"; \
	do sleep 1; done
.PHONY: mongo/prepare

.EXPORT_ALL_VARIABLES:
