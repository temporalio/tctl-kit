############################# Main targets #############################
# Install all tools, recompile proto files, run all possible checks and tests (long but comprehensive).
all: test
########################################################################


##### Variables ######
COLOR := "\e[1;36m%s\e[0m\n"

##### Test #####
test:
	@printf $(COLOR) "Running unit tests..."
	go test ./...
