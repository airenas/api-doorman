-include Makefile.options
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' 
.PHONY: help
#####################################################################################
test: 
	go test -v ./...
.PHONY: test
#####################################################################################
## build doorman-admin
build/doorman-admin: 
	cd deploy/doorman-admin && $(MAKE) clean dbuild
## run integration tests
test/integration: 
	cd testing/integration/cms && $(MAKE) test/integration clean
.PHONY: test/integration
#####################################################################################
generate: 
	go get github.com/petergtz/pegomock/...
	go generate ./...

build:
	cd cmd/doorman/ && go build .

run:
	cd cmd/doorman/ && go run . -c config.yml	

build-docker:
	cd deploy/doorman && $(MAKE) dbuild	

push-docker:
	cd deploy/doorman && $(MAKE) dpush

run-admin:
	cd cmd/doorman-admin/ && go run . -c config.yml	

build-docker-admin:
	cd deploy/doorman-admin && $(MAKE) dbuild	

push-docker-admin:
	cd deploy/doorman-admin && $(MAKE) dpush	

clean:
	rm -f cmd/audio-len/audio-len
	cd deploy/doorman && $(MAKE) clean
	cd deploy/doorman-admin && $(MAKE) clean

