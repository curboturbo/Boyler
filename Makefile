OS = linux
ARCH = amd_64

.PHONY: compile
compile:
	go build -o bin/boyler ./cmd/boyler
	go build -o bin/myrunc ./cmd/myrunc
	go build -o bin/daemon_boyler_$(OS) ./cmd/boylerd
	@echo "Binary files was created"


.PHONY: genproto
genproto:
	protoc --proto_path=proto \
		--go_out=internal/daemon/infrastructure/inbound/api/grpc/gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=internal/daemon/infrastructure/inbound/api/grpc/gen \
		--go-grpc_opt=paths=source_relative \
		proto/daemon.proto


.PHONY: prepare
prepare:
	-mkdir lib
	-mkdir lib/containers
	-mkdir lib/images
	-mkdir bin


.PHONY: clean
clean:
	-sudo ip link del boyler0