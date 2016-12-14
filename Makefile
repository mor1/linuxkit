all:
	$(MAKE) -C alpine

aufs:
	$(MAKE) AUFS=true all

alpine/initrd.img:
	$(MAKE) -C alpine initrd.img

alpine/initrd-test.img:
	$(MAKE) -C alpine

alpine/kernel/x86_64/vmlinuz64:
	$(MAKE) -C alpine/kernel x86_64/vmlinuz64

alpine/mobylinux-bios.iso:
	$(MAKE) -C alpine mobylinux-bios.iso

qemu: Dockerfile.qemu alpine/initrd.img alpine/kernel/x86_64/vmlinuz64
	tar cf - $^ | docker build -f Dockerfile.qemu -t mobyqemu:build -
	docker run -it --rm mobyqemu:build

qemu-iso: Dockerfile.qemuiso alpine/mobylinux-bios.iso
	tar cf - $^ | docker build -f Dockerfile.qemuiso -t mobyqemuiso:build -
	docker run -it --rm mobyqemuiso:build

hyperkit.git:
	git clone https://github.com/docker/hyperkit.git hyperkit.git

hyperkit.git/build/com.docker.hyperkit: hyperkit.git
	cd hyperkit.git && make

hyperkit: hyperkit.sh hyperkit.git/build/com.docker.hyperkit alpine/initrd.img alpine/kernel/x86_64/vmlinuz64
	sudo ./hyperkit.sh

test: Dockerfile.test alpine/initrd-test.img alpine/kernel/x86_64/vmlinuz64
	$(MAKE) -C alpine
	BUILD=$$( tar cf - $^ | docker build -f Dockerfile.test -q - ) && \
	[ -n "$$BUILD" ] && \
	echo "Built $$BUILD" && \
	touch test.log && \
	docker run --rm $$BUILD 2>&1 | tee -a test.log
	@cat test.log | grep -q 'Moby test suite PASSED'

TAG=$(shell git rev-parse HEAD)
STATUS=$(shell git status -s)
MOBYLINUX_TAG=alpine/mobylinux.tag
ifdef AUFS
AUFS_PREFIX=aufs-
endif
MEDIA_IMAGE=mobylinux/media:$(MEDIA_PREFIX)$(AUFS_PREFIX)$(TAG)
KERNEL_IMAGE=mobylinux/kernel:$(MEDIA_PREFIX)$(AUFS_PREFIX)$(TAG)
media: Dockerfile.media alpine/initrd.img alpine/kernel/x86_64/vmlinuz64 alpine/mobylinux-efi.iso
ifeq ($(STATUS),)
	tar cf - $^ alpine/mobylinux.efi alpine/kernel/x86_64/vmlinux alpine/kernel/x86_64/kernel-headers.tar | docker build -f Dockerfile.media -t $(MEDIA_IMAGE) -
	docker push $(MEDIA_IMAGE)
	[ -f $(MOBYLINUX_TAG) ]
	docker tag $(shell cat $(MOBYLINUX_TAG)) $(MEDIA_IMAGE)
	docker push $(MEDIA_IMAGE)
	tar cf - Dockerfile.kernel alpine/kernel/x86_64/vmlinuz64 | docker build -f Dockerfile.kernel -t $(KERNEL_IMAGE) -
	docker push $(KERNEL_IMAGE)
else
	$(error "git not clean")
endif

get:
ifeq ($(STATUS),)
	IMAGE=$$( docker create mobylinux/media:$(MEDIA_PREFIX)$(TAG) /dev/null ) && \
	mkdir -p alpine/kernel/x86_64 && \
	docker cp $$IMAGE:vmlinuz64 alpine/kernel/x86_64/vmlinuz64 && \
	docker cp $$IMAGE:vmlinux alpine/kernel/x86_64/vmlinux && \
	docker cp $$IMAGE:kernel-headers.tar alpine/kernel/x86_64/kernel-headers.tar && \
	docker cp $$IMAGE:initrd.img alpine/initrd.img && \
	docker cp $$IMAGE:mobylinux-efi.iso alpine/mobylinux-efi.iso && \
	docker cp $$IMAGE:mobylinux.efi alpine/mobylinux.efi && \
	docker rm $$IMAGE
else
	$(error "git not clean")
endif

.PHONY: clean

clean:
	$(MAKE) -C alpine clean
	rm -rf hyperkit.git disk.img
