all:
	$(MAKE) -C alpine/kernel
	$(MAKE) -C alpine

xhyve: all
	$(MAKE) -C xhyve run

qemu: all
	docker build -t mobyqemu:build .
	docker run -it mobyqemu:build

qemu-arm: Dockerfile.armhf
	$(MAKE) -C alpine/kernel arm
	$(MAKE) -C alpine arm
	docker build -f Dockerfile.armhf -t mobyarmqemu:build .
	docker run -it mobyarmqemu:build

clean:
	$(MAKE) -C alpine clean
	$(MAKE) -C xhyve clean
	docker images -q mobyqemu:build | xargs docker rmi
