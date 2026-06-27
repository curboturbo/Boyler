
ROOT=/home/tema/Boyler

.PHONY: clean-test
clean-test:
	-sudo umount -l /home/tema/Boyler/lib/containers/a/merged
	-sudo rm -rf /home/tema/Boyler/lib/containers/a
	-sudo rm -rf /home/tema/Boyler/lib/images/alpine/rootfs
	@echo "Alpine images and junk was cleaned"


.PHONY: compile
compile:
	go build -o bin/cobra ./cmd/boyler
	go build -o bin/myrunc ./cmd/myrunc
	@echo "Binary files was created"
