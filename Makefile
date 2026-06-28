
ROOT=/home/tema/Boyler

.PHONY: clean-test
clean-test:
	-sudo umount -l /home/tema/Boyler/lib/containers/a/merged # отмонируем merged от overlay
	-sudo rm -rf /home/tema/Boyler/lib/containers/a
	-sudo rm -rf /home/tema/Boyler/lib/images/alpine/rootfs
	@echo "Alpine images and junk was cleaned"


.PHONY: compile
compile:
	go build -o bin/cobra ./cmd/boyler
	go build -o bin/myrunc ./cmd/myrunc
	@echo "Binary files was created"


.PHONY: run
run:
	-export $(cat .env | xargs) # проброс окружение в дочерние потоки 
	sudo -E ./bin/cobra run # подхватывает окружение
	@echo "Test fun finished"


.PHONY: cond
cond:
	sudo lsns -p 123 # показывает namespaces конекртногоо PID
	sudo nsenter -t 82050 -m -u -i -n -p /bin/sh # выбрасывает в терминал контйенра по PID
	# заставляет отрыть bash в namepaces процесса
