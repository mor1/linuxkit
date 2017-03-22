# Moby

Moby, a toolkit for building custom minimal, immutable Linux distributions.

- Good, secure defaults included
- Everything is replaceable and customisable
- Immutable infrastructure applied to building Linux distributions
- Completely stateless, but persistent storage can be attached
- Easy tooling, with easy iteration
- Built with containers, for running containers
- Designed for building and running clustered applications, including but not limited to container orchestration such as Docker or Kubernetes
- Designed from the experience of building Docker Editions, but redesigned as a general purpose toolkit
- Designed to be managed by external tooling, such as [Infrakit](https://github.com/docker/infrakit) or similar tools
- Includes a set of longer term collaborative projects in various stages of development to innovate on kernel and userspace changes, particularly around security

## Getting Started

### Build

Simple build instructions: use `make` to build.
This will build the Moby customisation tool and a Moby initrd image.

If you already have a Go build environment and installed the source in your `GOPATH`
you can do `go install github.com/docker/moby/cmd/moby` to install the `moby` tool
instead, and then use `moby build moby.yaml` to build the example configuration.

#### Build requirements

- GNU `make`
- GNU or BSD `tar` (not `busybox` `tar`)
- Docker

### Booting and Testing

If you have a recent version of Docker for Mac installed you can use `moby run <name>` to execute the image you created with `moby build <name>.yaml`

The Makefile also specifies a number of targets:
- `make qemu` will boot up a sample Moby in qemu in a container
- on OSX: `make hyperkit` will boot up Moby in hyperkit
- `make test` or `make hyperkit-test` will run the test suite
- There are also docs for booting on [Google Cloud](docs/gcp.md)
- More detailed docs will be available shortly, for running single hosts and clusters.

## Customise

To customise, copy or modify the [`moby.yaml`](moby.yaml) to your own `file.yaml` or use on of the [examples](examples/) and then run `./bin/moby build file.yaml` to
generate its specified output. You can run the output with `./scripts/qemu.sh` or `./scripts/hyperkit.sh`, or on other
platforms.

### Yaml Specification

The Yaml format is loosely based on Docker Compose:

- `kernel` specifies a kernel Docker image, containing a kernel and a filesystem tarball, eg containing modules. `mobylinux/kernel` is built from `kernel/`
- `init` is the base `init` process Docker image, which is unpacked as the base system, containing `init`, `containerd`, `runc` and a few tools. Built from `base/init/`
- `system` are the system containers, executed sequentially in order. They should terminate quickly when done.
- `daemon` is the system daemons, which normally run for the whole time
- `files` are additional files to add to the image
- `outputs` are descriptions of what to build, such as ISOs.

For the images, you can specify the configuration much like Compose, with some changes, eg `capabilities` must be specified in full, rather than `add` and `drop`, and
there are no volumes only `binds`.

The config is liable to be changed, and there are missing features; full documentation will be available shortly.
