-include Makefile.options
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' | sort 
.PHONY: help
#####################################################################################
## invoke unit tests
test/unit: 
	go test -v -race ./...
.PHONY: test/unit
## code vet and lint
test/lint: 
	go vet ./...
	go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
	go mod tidy
.PHONY: test/lint
#####################################################################################
## build doorman-admin
docker/doorman-admin/build: 
	cd build/doorman-admin && $(MAKE) dbuild
.PHONY: docker/doorman-admin/build
## build doorman
docker/doorman/build: 
	cd build/doorman && $(MAKE) dbuild	
.PHONY: docker/doorman/build
## scan doorman-admin for vulnerabilities
docker/doorman-admin/scan:
	cd build/doorman-admin && $(MAKE) dscan
.PHONY: docker/doorman-admin/scan	
## scan doorman for vulnerabilities
docker/doorman/scan:
	cd build/doorman && $(MAKE) dscan		
.PHONY: docker/doorman/scan
## run integration tests
test/integration: 
	cd testing/integration/cms && ( $(MAKE) -j1 test/integration clean || ( $(MAKE) clean; exit 1; ))
.PHONY: test/integration
## run load tests - start services, do load tests, clean services
test/load: 
	cd testing/load && $(MAKE) start all clean	
.PHONY: test/load
#####################################################################################
## generate mock objects for test
generate: 
	go install github.com/petergtz/pegomock/...@latest
	go generate ./...
.PHONY: generate	
## push doorman-admin docker
docker/doorman-admin/push:
	cd build/doorman-admin && $(MAKE) dpush
.PHONY: docker/doorman-admin/push		
## push doorman docker

docker/doorman/push:
	cd build/doorman && $(MAKE) dpush		
.PHONY: docker/doorman/push

run:
	cd cmd/doorman/ && go run . -c config.yml	
run-admin:
	cd cmd/doorman-admin/ && go run . -c config.yml	

clean:
	go clean 
	go mod tidy -compat=1.17
