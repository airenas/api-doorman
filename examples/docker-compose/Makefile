-include .env
#####################################################################################
.mongo:
	mkdir -p $@

start: | .mongo
	docker-compose up -d

stop:
	docker-compose stop && docker-compose rm -f

clean: stop
	rm -rf .mongo
#####################################################################################
key?=u2OA9U4yLB
prj?=test
adm-key-list:
	curl -X GET http://localhost:$(ADMIN_PORT)/$(prj)/key-list | jq
adm-add:
	curl -i -X POST http://localhost:$(ADMIN_PORT)/$(prj)/key -d '{"key":"$(key)", "limit":500, "validTo":"2030-11-24T11:07:00Z"}'
adm-get:
	curl -X GET http://localhost:$(ADMIN_PORT)/$(prj)/key/$(key)?full=1 | jq
adm-update:
	curl -X PATCH http://localhost:$(ADMIN_PORT)/$(prj)/key/$(key) -d '{"limit":1110, "disabled":false}'

#####################################################################################
test-public:
	curl -i -X POST http://localhost:$(DOORMAN_PORT)/public -d '{"text": "public text data"}'

test-private:
	curl -i -X POST http://localhost:$(DOORMAN_PORT)/private -d '{"text": "private text data"}'

test-private-key:
	curl -i -X POST http://localhost:$(DOORMAN_PORT)/private?key=$(key) -d '{"text": "private text data with key"}'	


