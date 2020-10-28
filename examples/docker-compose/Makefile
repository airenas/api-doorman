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
adm-key-list:
	curl -X GET http://localhost:$(ADMIN_PORT)/key-list | jq
adm-add:
	curl -i -X POST http://localhost:8001/key -d '{"limit":500, "validTo":"2030-11-24T11:07:00Z"}'
adm-get:
	curl -X GET http://localhost:8001/key/$(key)?full=1 | jq
adm-update:
	curl -X PATCH http://localhost:8001/key/$(key) -d '{"limit":1110, "disabled":false}'




