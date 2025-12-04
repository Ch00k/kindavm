.PHONY: lint test test-verbose test-one test-ci build build-native run release-patch release-minor release-major

.EXPORT_ALL_VARIABLES:

KINDAVMD_EXECUTABLE_FILENAME ?= kindavmd
KINDAVMD_BUILD_ARTIFACTS_DIR ?= dist
KINDAVMD_VERSION ?= dev

lint:
	golangci-lint run --fix
	uv run ruff format .
	uv run ruff check --fix .
	uv run mypy --config-file pyproject.toml .

test:
	gotestsum --format testname -- ./...

test-verbose:
	gotestsum --format standard-verbose -- -v -count=1 ./...

test-one:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	gotestsum --format standard-verbose -- -v -count=1 -run "^$(TEST)$$" ./...

test-ci:
	go run gotest.tools/gotestsum@latest --format testname -- -race "-coverprofile=coverage.txt" "-covermode=atomic" ./...

build:
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X main.Version=${KINDAVMD_VERSION}" -o ./${KINDAVMD_BUILD_ARTIFACTS_DIR}/${KINDAVMD_EXECUTABLE_FILENAME} ./cmd/kindavmd

build-native:
	go build -trimpath -ldflags="-s -w -X main.Version=${KINDAVMD_VERSION}" -o ./${KINDAVMD_BUILD_ARTIFACTS_DIR}/${KINDAVMD_EXECUTABLE_FILENAME}-native ./cmd/kindavmd

run:
	go run ./cmd/kindavmd

release-patch:
	./release.sh patch

release-minor:
	./release.sh minor

release-major:
	./release.sh major
