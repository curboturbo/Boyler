
ROOT=/home/tema/Boyler

.PHONY: clean
clean-test:
	-sudo umount -l /home/tema/Boyler/lib/containers/8e6ff240-a1a5-4642-a58e-1a53b3222ee0/merged # отмонируем merged от overlay
	-sudo rm -rf /home/tema/Boyler/lib/containers/8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	-sudo rm -rf /home/tema/Boyler/lib/images/python-alpine/rootfs
	-sudo rm -rf /var/run/myrunc/8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	-sudo rm /var/log/boyler_daemon.log
	-sudo rm /tmp/daemon_grpc.sock
	-sudo ip link del boyler0
	@echo "Alpine images and junk was cleaned"


OS = linux
ARCH = amd_64

.PHONY: compile
compile:
	go build -o bin/boyler ./cmd/boyler
	go build -o bin/myrunc ./cmd/myrunc
	go build -o bin/daemon_boyler_$(OS) ./cmd/boylerd
	@echo "Binary files was created"

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: init
init:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler init
	@echo "Test run finished"

.PHONY: create
create:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler create python-alpine --name HELLO_WORLD
	@echo "Test run finished"


.PHONY: pull
pull:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler pull python:alpine
	@echo "Test run finished"

.PHONY: stop
stop:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler stop 8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	@echo "Test run finished"

.PHONY: remove
remove:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler remove 8e6ff240-a1a5-4642-a58e-1a53b3222ee0



.PHONY: start
start:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler start 8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	@echo "Test run finished"


.PHONY: inspect
inspect:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler inspect 8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	@echo "Test run finished"

.PHONY: exec
exec:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler exec -it 8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	@echo "Test run finished"

.PHONY: ps
ps:
	export $$(cat $(ROOT_DIR).env | xargs) && sudo -E $(ROOT_DIR)bin/boyler ps
	@echo "Test run finished"


.PHONY: cond
cond:
	sudo lsns -p 123 # показывает namespaces конекртногоо PID
	sudo nsenter -t 23506 -m -u -i -n -p -r -w /bin/sh # выбрасывает в терминал контйенра по PID
	# заставляет отрыть bash в namepaces процесса

.PHONY: genproto
genproto:
	protoc --proto_path=proto \
		--go_out=internal/daemon/infrastructure/inbound/api/grpc/gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=internal/daemon/infrastructure/inbound/api/grpc/gen \
		--go-grpc_opt=paths=source_relative \
		proto/daemon.proto


# sudo rm -f /etc/resolv.conf
# echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf 
# setup dns for using VPN in host machine


.PHONY: default
default:
	-mkdir lib
	-mkdir lib/containers
	-mkdir lib/images
	-mkdir bin