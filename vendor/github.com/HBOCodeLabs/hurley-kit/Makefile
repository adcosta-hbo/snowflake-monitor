include Hurley-Make/Makefile.docker
include Hurley-Make/Makefile.golang
include Hurley-Make/Makefile.common

GO_PROJECT_NAME=github.com/HBOCodeLabs/hurley-kit
GOLANG_VERSION := 1.12.5
ARTIFACT_MAINTAINER=content-services

info: ## Print project Name and Version
	${INFO} "Project: $(ARTIFACT_NAME):$(ARTIFACT_VERSION)"
	${INFO} "Golang: $(GOLANG_VERSION)"
	${OK}

###################################################################################################
##     Building project
###################################################################################################
dependencies-clean: ## Removes vendor dependencies for this project
	${INFO}"Removing vendor dependencies"
	@rm -rf vendor || true
	${OK}

# Generate vendor dependency for testing in docker only
dependencies-prepare: ## Prepare build dependencies for this project
	${INFO}"Preparing all third-party dependencies"
	@mkdir -p vendor
	docker run --rm \
		-v ${PROJECT_DIR}:/go/src/${GO_PROJECT_NAME} \
		-w /go/src/${GO_PROJECT_NAME} \
		${GOLANG_DOCKER_ENV} \
    ${GOLANG_IMAGE} \
		sh -c 'go get -v -t -d ${GO_PROJECT_NAME}/... && govendor init && govendor fetch -v +external && govendor list'
	${OK}

	${INFO}"retrieving non-HEAD dependencies"

# Fixes the fact that the jaeger client doesn't currently build
# against the latest version of Thrift and go.uuid:
# https://github.com/jaegertracing/jaeger-client-go/issues/177
# https://github.com/jaegertracing/jaeger-client-go/issues/292
	docker run --rm \
		-v ${PROJECT_DIR}:/go/src/${GO_PROJECT_NAME} \
		-w /go/src/${GO_PROJECT_NAME} \
		${GOLANG_DOCKER_ENV} \
    ${GOLANG_IMAGE} \
		sh -c 'govendor fetch github.com/apache/thrift/lib/go/thrift@0dd823580c78a79ae9696eb9b3650e400fff140f && govendor fetch github.com/satori/go.uuid@v1.2.0'
	${OK}

	@$(call go-owner,$(GO_PROJECT_NAME),vendor)

compile: ## Compile go files
	${INFO}"compiling all packages"
	@$(call go-compile,$(GO_PROJECT_NAME))
	${OK}

build: ## Build and Test this project
	${INFO}"Building: $(call HLIGHT,$(ARTIFACT_NAME))"
	@make -s dependencies-prepare
	@make -s compile
	@make -s cover
.PHONY: build

cover:
	$(call go-cover,$(GO_PROJECT_NAME)) ./...
	$(call go-cobertura,$(GO_PROJECT_NAME))
	${OK}
	
.PHONY: cover

# All go files minus vendor
test: ## Run tests
	@$(call go-test,$(GO_PROJECT_NAME))
	${OK}

release-prepare: ## Prepares the release for this project
	${INFO}"Preparing the release"
	@make -s version-update
