.PHONY: tag push

ORG=mor1

ARCH=x86
BASE=alpine:3.5
IMAGE=alpine-build-kernel
PACKAGES=/lib/apk/db/installed

ARCH=arm64
BASE=debian:stretch
IMAGE=alpine-build-kernel-$(ARCH)
PACKAGES=/var/lib/dpkg/status
TRUST=0

default: push

hash:
	DOCKER_CONTENT_TRUST=$(TRUST) docker pull $(BASE)
	tar cf - Dockerfile.$(ARCH) | docker build --no-cache -f Dockerfile.$(ARCH) -t $(IMAGE):build -
	docker run --rm $(IMAGE):build sha1sum $(PACKAGES) | sed 's/ .*//' > hash

push: hash
	docker pull $(ORG)/$(IMAGE):$(shell cat hash) || \
		(docker tag $(IMAGE):build $(ORG)/$(IMAGE):$(shell cat hash) && \
		 docker push $(ORG)/$(IMAGE):$(shell cat hash))
	docker rmi $(IMAGE):build
	rm -f hash

tag: hash
	docker pull $(ORG)/$(IMAGE):$(shell cat hash) || \
		docker tag $(IMAGE):build $(ORG)/$(IMAGE):$(shell cat hash)
	docker rmi $(IMAGE):build
	rm -f hash

clean:
	rm -f hash

.DELETE_ON_ERROR:
