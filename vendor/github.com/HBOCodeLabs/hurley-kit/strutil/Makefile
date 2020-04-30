include Hurley-Make/Makefile.docker
include Hurley-Make/Makefile.golang
include Hurley-Make/Makefile.common

GO_PROJECT_NAME=HBOCodeLabs/strutil
GOLANG_VERSION := 1.9.0
ARTIFACT_MAINTAINER=platform-infrastructure

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
	${INFO}"Preparing vendor dependencies"
	@mkdir -p vendor
	@$(call go-vendor,$(GO_PROJECT_NAME))
	@$(call go-owner,$(GO_PROJECT_NAME),vendor)
	${OK}

compile: ## Compile go files
	${INFO}"building binary: $(call HLIGHT,$@)"
	@$(call go-compile,$(GO_PROJECT_NAME))
	${OK}

build: ## Build and Test this project
	${INFO}"Building: $(call HLIGHT,$(ARTIFACT_NAME))"
	@make -s dependencies-prepare
	@make -s test
.PHONY: build

# All go files minus vendor
test: ## Run tests
	@$(call go-test,$(GO_PROJECT_NAME))
	${OK}

###################################################################################################
##     Release
###################################################################################################
release-prepare: ## Prepares the release for this project
	${INFO}"Preparing the release"
	@make -s version-update