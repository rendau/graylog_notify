.DEFAULT_GOAL := build

CMD_PATH = cmd
BINARY_NAME = svc
BUILD_PATH = $(CMD_PATH)/build

build:
	mkdir -p $(BUILD_PATH)
	CGO_ENABLED=0 go build -o $(BUILD_PATH)/$(BINARY_NAME) $(CMD_PATH)/main.go

clean:
	rm -rf $(BUILD_PATH)
