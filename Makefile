# Define pkgs, run, and cover vairables for test so that we can override them in
# the terminal more easily.
pkgs := $(shell go list ./...)
run := .
count := 1

## help: Show this help message
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

## clean: clean testcache
clean:
	@echo "--> Clearing testcache"
	@go clean --testcache
.PHONY: clean

## cover: generate to code coverage report.
cover:
	@echo "--> Generating Code Coverage"
	@go install github.com/ory/go-acc@latest
	@go-acc -o coverage.txt $(pkgs)
.PHONY: cover

## deps: Install dependencies
deps:
	@echo "--> Installing dependencies"
	@go mod download
#	@make proto-gen
	@go mod tidy
.PHONY: deps

## lint: Run linters golangci-lint and markdownlint.
lint: vet
	@echo "--> Running golangci-lint"
	@golangci-lint run
	@echo "--> Running markdownlint"
	@markdownlint --config .markdownlint.yaml '**/*.md'
	@echo "--> Running yamllint"
	@yamllint --no-warnings . -c .yamllint.yml

.PHONY: lint

## fmt: Run fixes for linters. Currently only markdownlint.
fmt:
	@echo "--> Formatting markdownlint"
	@markdownlint --config .markdownlint.yaml '**/*.md' -f
	@echo "--> Formatting go"
	@golangci-lint run --fix
.PHONY: fmt

## vet: Run go vet
vet: 
	@echo "--> Running go vet"
	@go vet $(pkgs)
.PHONY: vet

## test: Running unit tests
test: vet
	@echo "--> Running unit tests"
	@go test -v -race -covermode=atomic -coverprofile=coverage.txt $(pkgs) -run $(run) -count=$(count)
.PHONY: test