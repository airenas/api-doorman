-include .env
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' 
#####################################################################################
## start all containers
start: 
	docker-compose up -d mongo
	$(MAKE) mongo-prepare 
	docker-compose up -d doorman-admin
## invoke integration tests
test/integration: start 
	docker-compose up integration-tests
## clean everything, stops docker containers and removes them
clean:
	docker-compose stop && docker-compose rm -f
#####################################################################################
mongo-prepare:
	until docker-compose exec -T mongo /usr/bin/mongo -u $(MONGO_USER) -p $(MONGO_PASS) --eval "rs.initiate()"; \
	do sleep 1; done
