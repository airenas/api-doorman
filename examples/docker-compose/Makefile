-include .env
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' 
#####################################################################################
.mongo:
	mkdir -p $@

## start all test containers
start: | .mongo
	docker-compose up -d mongo
	$(MAKE) mongo-prepare 
	docker-compose up -d

## stop containers
stop:
	docker-compose stop && docker-compose rm -f

## clean everything. Perhaps you should run 'sudo make clean'
clean: stop
	rm -rf .mongo
#####################################################################################
key?=u2OA9U4yLB
prj?=test
## return all keys 
test/adm/key-list:
	curl -X GET http://localhost:$(ADMIN_PORT)/$(prj)/key-list | jq .
## add new key 
test/adm/add:
	curl -i -X POST http://localhost:$(ADMIN_PORT)/$(prj)/key \
		-H "content-type:application/json" \
		-d '{"key":"$(key)", "limit":500, "validTo":"2030-11-24T11:07:00Z"}'
## get info about key 
test/adm/get:
	curl -X GET http://localhost:$(ADMIN_PORT)/$(prj)/key/$(key)?full=1 | jq
## sample on how to update key
test/adm/update:
	curl -X PATCH http://localhost:$(ADMIN_PORT)/$(prj)/key/$(key) \
		-H "content-type:application/json" \
		-d '{"limit":1110, "disabled":false}'

#####################################################################################
## invoke public testing endpoint. No auth here
test/api/public:
	curl -i -X POST http://localhost:$(DOORMAN_PORT)/public -d '{"text": "public text data"}'
## invoke private API. Auth on IP address
test/api/private:
	curl -i -X POST http://localhost:$(DOORMAN_PORT)/private -d '{"text": "private text data"}'
## invoke private API with a key provided
test/api/key:
	curl -i -X POST http://localhost:$(DOORMAN_PORT)/private -H "Authorization: Key $(key)" \
	-d '{"text": "private text data with key"}'	

########################################################################################################
mongosh:
	docker-compose exec mongo /usr/bin/mongo -u $(MONGO_USER) -p $(MONGO_PASS)

mongo-prepare:
	until docker-compose exec mongo /usr/bin/mongo -u $(MONGO_USER) -p $(MONGO_PASS) --eval "rs.initiate()"; \
	do sleep 1; done


mongo-update-script:
	docker-compose exec mongo sh -c \
	'printf "use test\ndb.log.update({}, {\$$rename:{\"quotavalue\":\"quotaValue\", \"responsecode\":\"responseCode\"}}, false, true)" | /usr/bin/mongo -u $(MONGO_USER) -p $(MONGO_PASS)'

#####################################################################################
## test cms integration - add new key
test/cms/add-key:
	curl -i -X POST http://localhost:$(ADMIN_PORT)/key \
		-H "content-type:application/json" \
		-d '{"service":"test", "id": "123", "operationID": "op123", "credits": 100}'

## test cms integration - return key
test/cms/key:
	curl -i -X GET http://localhost:$(ADMIN_PORT)/key/123?returnKey=1

## test cms integration - add credits
test/cms/add-credits:
	curl -i -X PATCH http://localhost:$(ADMIN_PORT)/key/123/credits	\
	-H "content-type:application/json" \
	-d '{"operationID": "op124", "credits": 100}'

## test cms integration - return usage
test/cms/usage:
	curl -i -X GET http://localhost:$(ADMIN_PORT)/key/123/usage?full=1&&from=2021-01-02T00:00:00Z 
