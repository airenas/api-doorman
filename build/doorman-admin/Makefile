#####################################################################################
service=airenas/api-doorman-admin
version?=dev
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild: $(dist_dir)/$(executable_name)
	cd ../../ && docker buildx build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/doorman-admin/Dockerfile .

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)		
#####################################################################################
.PHONY:
	dbuild dpush dscan	
