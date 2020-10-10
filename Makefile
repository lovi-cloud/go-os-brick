.DEFAULT_GOAL := help

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## test by gotgt (support multi platform)
	act -j test -P ubuntu-latest=whywaita/iscsi-client

test-openicsi: ## test by real open-iscsi (only support on linux)
	go test -tags=host -v ./...