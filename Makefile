-include Makefile.options

test: 
	go test ./...

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

