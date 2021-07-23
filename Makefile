APP = etamu-file2tar
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...

BIN ?= $(OUTPUT_DIR)/$(APP)

GO_FLAGS ?= -v -mod=vendor
GO_TEST_FLAGS ?= -race -cover

TEST_BASE_DIR ?= /var/tmp/$(APP)

ARGS ?=

.PHONY: $(BIN)
$(BIN):
	go build $(GO_FLAGS) -o $(BIN) $(CMD)

build: $(BIN)

install: build
	go install $(GO_FLAGS) $(CMD)

clean:
	rm -rf "$(OUTPUT_DIR)" || true

clean-temp-dirs:
	rm -rf $(TEST_BASE_DIR) || true

mk-temp-dirs:
	mkdir -p $(TEST_BASE_DIR)/{base,work} || true

start: mk-temp-dirs
	BASE_DIR="$(TEST_BASE_DIR)/base" \
	WORK_DIR="$(TEST_BASE_DIR)/work" \
		go run $(GO_FLAGS) $(CMD) $(ARGS)

test: test-unit

.PHONY: test-unit
test-unit:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(CMD) $(PKG) $(ARGS)