#####################################################################################
service=airenas/api-doorman-migrator
version?=dev
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild: $(dist_dir)/$(executable_name)
	cd ../../ && docker buildx build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/doorman-migrator/Dockerfile .

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)		
#####################################################################################
.PHONY:
	dbuild dpush dscan
