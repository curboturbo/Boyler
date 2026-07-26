
ROOT=/home/tema/Boyler

.PHONY: clean
clean-test:
	-sudo umount -l /home/tema/Boyler/lib/containers/8e6ff240-a1a5-4642-a58e-1a53b3222ee0/merged # отмонируем merged от overlay
	-sudo rm -rf /home/tema/Boyler/lib/containers/8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	-sudo rm -rf /home/tema/Boyler/lib/images/alpine/rootfs
	-sudo rm -rf /var/run/myrunc/8e6ff240-a1a5-4642-a58e-1a53b3222ee0
	@echo "Alpine images and junk was cleaned"


.PHONY: compile
compile:
	go build -o bin/cobra ./cmd/boyler
	go build -o bin/myrunc ./cmd/myrunc
	@echo "Binary files was created"


.PHONY: run
run:
	-export $$(cat .env | xargs) && sudo -E ./bin/cobra run
	@echo "Test fun finished"


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