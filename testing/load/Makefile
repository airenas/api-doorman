-include .env
-include ../../version
#####################################################################################
DOCKER_IP=$(shell echo $$(ip addr show docker0 | grep -Po 'inet \K[\d.]+'))
ADMIN_VERSION=$(admin_version)
DOORMAN_VERSION=$(doorman_version)
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' 
.PHONY: help
#####################################################################################
## start all containers for testing
start: 
	docker compose up -d dbmigration doorman-admin doorman test-api 
	docker compose up -d doorman --wait doorman --timeout 10
.PHONY: start
## clean everything, stops docker containers and removes them
clean:
	docker compose down 
.PHONY: clean	
#####################################################################################
psql:
	docker compose exec -e PGPASSWORD=postgres postgres psql -U postgres -h localhost -d doorman
#####################################################################################
vu?=10
req?=1000

info:
	@echo IP: $(DOCKER_IP)

load: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
		--env EXPECTED_REQ=$(req) \
	    -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_OK.js
.PHONY: load

load-fail: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
	   --env EXPECTED_REQ=$(req) \
	   -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_Fail.js		
.PHONY: load-fail

load-restore: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
	   -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_Restore.js	
.PHONY: load-restore

load-discount: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
		--env EXPECTED_REQ=$(req) \
	    -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_Discount.js
.PHONY: load-discount
## run all load tests suite
all: load load-fail load-restore load-discount
.PHONY: all

.EXPORT_ALL_VARIABLES:
