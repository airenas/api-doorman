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
build/doorman-admin: 
	cd build/doorman-admin && $(MAKE) clean dbuild
.PHONY: build/doorman-admin
## build doorman
docker/doorman/build: 
	cd build/doorman && $(MAKE) dbuild	
.PHONY: docker/doorman/build
## scan doorman for vulnerabilities
docker/doorman/scan:
	cd build/doorman && $(MAKE) dscan	
.PHONY: docker/doorman/scan
## run integration tests
test/integration: 
	cd testing/integration/cms && $(MAKE) test/integration clean
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

run:
	cd cmd/doorman/ && go run . -c config.yml	

push-docker:
	cd deploy/doorman && $(MAKE) dpush

run-admin:
	cd cmd/doorman-admin/ && go run . -c config.yml	

push-docker-admin:
	cd deploy/doorman-admin && $(MAKE) dpush	

clean:
	go clean 
	go mod tidy -compat=1.17
	cd build/doorman-admin && $(MAKE) clean
