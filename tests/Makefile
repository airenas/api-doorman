DOCKER_IP=$(shell echo $$(ip addr show docker0 | grep -Po 'inet \K[\d.]+'))

vu?=10
req?=1000

info:
	@echo IP: $(DOCKER_IP)

load: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
		--env EXPECTED_REQ=$(req) \
	    -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_OK.js

load-fail: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
	   --env EXPECTED_REQ=$(req) \
	   -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_Fail.js		

load-restore: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
	   -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_Restore.js	

all: load load-fail load-restore

.PHONY:
	load load-restore load-fail