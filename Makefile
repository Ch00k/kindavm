.PHONY: lint test test-verbose test-one test-ci build build-native run package release-patch release-minor release-major

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

package: build
	@echo "Creating release package..."
	@mkdir -p $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package
	@cp $(KINDAVMD_BUILD_ARTIFACTS_DIR)/$(KINDAVMD_EXECUTABLE_FILENAME) $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/kindavmd
	@cp init_hid.sh $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@cp init_hdmi.sh $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@cp uninstall.sh $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@cp kindavm-init.service $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@cp kindavmd.service $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@cp tools/edidmod $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@cp tools/ustreamer $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/
	@mkdir -p $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/configs
	@cp configs/boot.conf $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/configs/
	@cp configs/edid.hex $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/configs/
	@cp configs/hid_report_desc.bin $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/configs/
	@cp configs/modules.conf $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package/configs/
	@cd $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package && tar -czf ../kindavm-linux-arm64.tar.gz *
	@rm -rf $(KINDAVMD_BUILD_ARTIFACTS_DIR)/package
	@echo "Package created: $(KINDAVMD_BUILD_ARTIFACTS_DIR)/kindavm-linux-arm64.tar.gz"

release-patch:
	./release.sh patch

release-minor:
	./release.sh minor

release-major:
	./release.sh major
