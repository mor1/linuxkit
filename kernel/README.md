Linux kernel builds, based on mostly-vanilla upstream Linux kernels.
See [../docs/kernel-patches.md] for how the local patches in `patches-*`
are maintained.

The build is mostly silent. A local build can be run via `make`. To view
the output use `docker log -f <containerid>`. The build creates multiple
containers, so multiple invocations may be necessary. To view the full build
output one may also invoke `docker build .` and then copy the build artefacts
from the image afterwards.

To build with various debug options enabled, build the kernel with
`make DEBUG=1`. The options enabled are listed in `kernel_config.debug`.
This allocates a significant amount of memory on boot and you may need to
adjust the kernel config on some systems. Specifically:

```diff
--- a/alpine/kernel/kernel_config
+++ b/alpine/kernel/kernel_config
@@ -415,8 +415,8 @@ CONFIG_DMI=y
 # CONFIG_CALGARY_IOMMU is not set
 CONFIG_SWIOTLB=y
 CONFIG_IOMMU_HELPER=y
-CONFIG_MAXSMP=y
-CONFIG_NR_CPUS=8192
+CONFIG_MAXSMP=n
+CONFIG_NR_CPUS=8
 # CONFIG_SCHED_SMT is not set
 CONFIG_SCHED_MC=y
 # CONFIG_PREEMPT_NONE is not set
```

## Notes

Current kernel build workflow is:
  * `make -C base/alpine-build-kernel` to build the Alpine builder image
      * Default is `mobylinux/alpine-build-kernel`
      * I've pushed `mor1/alpine-build-kernel-arm64`, a Debian builder for ARM64
  * `make -C kernel kernel.tag` to build the kernel continer
      * Uses `tar` to pass through the necessary kernel patchsets and config
        files to invoke the build.
  * `make -C kernel bzImage` to invoke `tar` to extract contents from kernel
    container
  * `make -C kernel image` to build container containing built kernel
  * `make -C kernel push` to push it

We need to support multiple kernel versions (patched and vanilla) as well as
multiple architectures (x86_64 and ARM64 to start).

For kernel build, we (will) need to specify
  * `ARCH`
      * `arm64` or `x86_64` (default)
  * `KERNEL_VERSION` deriving `IMAGE_VERSION` and `IMAGE_MAJOR_VERSION`
      * `4.4` or `4.9` (default)
  * `DEBUG`
      * `1` or `0` (default)
  * `PATCHED`
      * Not currently supported

These need to be plumbed through to give us:
  * Kernel builder base container
      * Specify `Dockerfile` in `base/alpine-build-kernel`
      * Publish to alternative tag (?)
  * Kernel configurations, composed of parts
      * Common
      * Common debug-only
      * Version specific
      * Architecture specific
      * Moby patchsets
  * Target-specific kernel build container
      * Need to pass through different `DEPS` in `kernel/Makefile`

Questions:
  * Better to setup cross-compile rather than rely on QEMU,
    per <https://github.com/mor1/xen-arm-builder/blob/master/linux.sh>?
  * Easier to put each `Dockerfile` in own subdir
      * Eg., `$(ARCH)/$(IMAGE_VERSION)` with patches split by `IMAGE_VERSION`?
  * Relationship between `KERNEL_VERSION`, `IMAGE_VERSION` and `PATCHED`?
  * To build kernel builder, flip tag in `FROM` vs use separate
    `Dockerfile`/`Dockerfile.arm64`?
  * How many kernel versions and patchsets will we ever want live at once?
  * Can we overload image tags until we have multiarch support?
