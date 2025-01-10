-include Makefile.options
#####################################################################################
MIGRATIONS_DIR := ./db/migrations
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
	go test -race ./internal/... ./cmd/...
.PHONY: test/unit
#####################################################################################
## code vet and lint
test/lint: 
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run -v --enable gofmt --enable misspell --enable zerologlint --timeout 5m  ./...
.PHONY: test/lint
#####################################################################################
## build doorman-admin
docker/doorman-admin/build: 
	cd build/doorman-admin && $(MAKE) dbuild
.PHONY: docker/doorman-admin/build
## build doorman-admin-migration
docker/api-doorman-dbmigration/build: 
	cd build/api-doorman-dbmigration && $(MAKE) dbuild
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
## build all docker images
docker/build/all: docker/doorman-admin/build docker/doorman/build docker/api-doorman-dbmigration/build
.PHONY: docker/build/all
#####################################################################################
## run integration tests
test/integration: test/integration/cms test/integration/doorman 
.PHONY: test/integration
## run cms integration tests
test/integration/cms: 
	cd testing/integration/cms && ( $(MAKE) -j1 test/integration clean || ( $(MAKE) clean; exit 1; ))
.PHONY: test/integration/cms
## run doorman integration tests
test/integration/doorman: 
	cd testing/integration/doorman && ( $(MAKE) -j1 test/integration clean || ( $(MAKE) clean; exit 1; ))
.PHONY: test/integration/doorman
## run load tests - start services, do load tests, clean services
test/load: 
	cd testing/load && $(MAKE) start all clean	
.PHONY: test/load
#####################################################################################
## generate mock objects for test
generate:
	go install github.com/petergtz/pegomock/v4/pegomock
	go generate ./...
.PHONY: generate	
#####################################################################################
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
#####################################################################################
install/migrate:
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
.PHONY: install-migrate
$(MIGRATIONS_DIR):
	mkdir -p $@
migrate/new: install/migrate | $(MIGRATIONS_DIR)
	@$(if $(strip $(migration_name)),echo "Creating = $(migration_name)",echo No migration_name && exit 1)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(migration_name)
.PHONY: migrate/new
migrate/up: install/migrate 
	migrate -path=$(MIGRATIONS_DIR) -database="$(DB_DSN)" up
.PHONY: migrate/up
# rollback 1 migration
migrate/down: install/migrate
	migrate -path=$(MIGRATIONS_DIR) -database="$(DB_DSN)" down 1
.PHONY: migrate/down
#####################################################################################
clean:
	go clean 
	go mod tidy -compat=1.19
