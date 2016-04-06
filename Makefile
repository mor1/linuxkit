all:
	$(MAKE) -C alpine/kernel
	$(MAKE) -C alpine

xhyve: all
	$(MAKE) -C xhyve run

qemu: all
	docker build -f Dockerfile.qemu -t mobyqemu:build .
	docker run --rm mobyqemu:build

qemu-iso: all
	$(MAKE) -C alpine mobylinux.iso
	docker build -f Dockerfile.qemuiso -t mobyqemuiso:build .
	docker run --rm mobyqemuiso:build

arm:
	$(MAKE) -C alpine/kernel arm
	$(MAKE) -C alpine arm

qemu-arm: Dockerfile.qemu.armhf arm
	docker build -f Dockerfile.qemu.armhf -t mobyarmqemu:build .
	docker run --rm mobyarmqemu:build

.PHONY: clean

clean:
	$(MAKE) -C alpine clean
	$(MAKE) -C xhyve clean
	docker images -q mobyqemu:build | xargs docker rmi -f
	docker images -q justincormack/remora | xargs docker rmi -f
