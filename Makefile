############################# Main targets #############################
# Install all tools, recompile proto files, run all possible checks and tests (long but comprehensive).
all: clean
########################################################################


##### Variables ######
ifndef GOOS
GOOS := $(shell go env GOOS)
endif

ifndef GOARCH
GOARCH := $(shell go env GOARCH)
endif

COLOR := "\e[1;36m%s\e[0m\n"

##### Test #####
TEST_TIMEOUT := 20m
ALL_SRC         := $(shell find . -name "*.go")
TEST_DIRS       := $(sort $(dir $(filter %_test.go,$(ALL_SRC))))

ifdef TEST_TAG
override TEST_TAG := -tags $(TEST_TAG)
endif

test: clean-test-results
	@printf $(COLOR) "Running unit tests..."
	$(foreach TEST_DIRS,$(TEST_DIRS),\
		@go test $(TEST_DIRS) -timeout=$(TEST_TIMEOUT) $(TEST_TAG) -race \
	$(NEWLINE))

clean-test-results:
	@rm -f test.log
	@go clean -testcache
