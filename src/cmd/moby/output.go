package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	bios = "mobylinux/mkimage-iso-bios:489b1f054a77a8f379d0bfc6cd91639b4db6b67c@sha256:0f058951aac4367d132682aa19eeb5cdcb05600a5d51fe5d0fcbd97b03ae4f87"
	efi  = "mobylinux/mkimage-iso-efi:d86021840d2180422bd2f59dcff2fcfb5aea6ad1@sha256:00a6dc21073a24763bc667cadd90c42b69cd69579f8036d6d794b42cbb583142"
	gce  = "mobylinux/mkimage-gce:2039be4e39e855d1845aee188e266bba3f1d2eed@sha256:e12f76003fd9eaa0c6f39f149db5998cf56de42539b989c994893c8344ca69c0"
	qcow = "mobylinux/mkimage-qcow:9b3632f111675898ed3a22ac71897e735b5a8364@sha256:2132cf3fb593d65f09c8d109d40e1fad138d81485d4750fc29a7f54611d78d35"
	vhd  = "mobylinux/mkimage-vhd:73c80e433bf717578c507621a84fd58cec27fe95@sha256:0ae1eda2d6592f309977dc4b25cca120cc4e2ee2cc786e88fdc2761c0d49cb14"
	vmdk = "mobylinux/mkimage-vmdk:1de360a30f3ac6a91d4eae1ae4611cea4b82f22a@sha256:d7e65edc6dd88f6e12dcb0d749d4c7e5793d1250e548b58c105dbfd082940787"
)

func outputs(m *Moby, base string, bzimage []byte, initrd []byte) error {
	log.Debugf("output: %s %s", m.Outputs, base)
	for _, o := range m.Outputs {
		switch o.Format {
		case "kernel+initrd":
			err := outputKernelInitrd(base, bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "iso-bios":
			err := outputISO(bios, base+".iso", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "iso-efi":
			err := outputISO(efi, base+"-efi.iso", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "gce-img":
			err := outputImg(gce, base+".img.tar.gz", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "gce-storage":
			err := outputImg(gce, base+".img.tar.gz", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
			if o.Bucket == "" {
				return fmt.Errorf("No bucket specified for GCE output")
			}
			err = uploadGS(base+".img.tar.gz", o.Project, o.Bucket, o.Public)
			if err != nil {
				return fmt.Errorf("Error copying to Google Storage: %v", err)
			}
		case "gce":
			err := outputImg(gce, base+".img.tar.gz", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
			if o.Bucket == "" {
				return fmt.Errorf("No bucket specified for GCE output")
			}
			err = uploadGS(base+".img.tar.gz", o.Project, o.Bucket, o.Public)
			if err != nil {
				return fmt.Errorf("Error copying to Google Storage: %v", err)
			}
			err = imageGS(base, o.Project, "https://storage.googleapis.com/"+o.Bucket+"/"+base+".img.tar.gz", o.Family, o.Replace)
			if err != nil {
				return fmt.Errorf("Error creating Google Compute Image: %v", err)
			}
		case "qcow", "qcow2":
			err := outputImg(qcow, base+".qcow2", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "vhd":
			err := outputImg(vhd, base+".vhd", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "vmdk":
			err := outputImg(vmdk, base+".vmdk", bzimage, initrd, m.Kernel.Cmdline)
			if err != nil {
				return fmt.Errorf("Error writing %s output: %v", o.Format, err)
			}
		case "":
			return fmt.Errorf("No format specified for output")
		default:
			return fmt.Errorf("Unknown output type %s", o.Format)
		}
	}
	return nil
}

func tarInitrdKernel(bzimage, initrd []byte) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	hdr := &tar.Header{
		Name: "bzImage",
		Mode: 0600,
		Size: int64(len(bzimage)),
	}
	err := tw.WriteHeader(hdr)
	if err != nil {
		return buf, err
	}
	_, err = tw.Write(bzimage)
	if err != nil {
		return buf, err
	}
	hdr = &tar.Header{
		Name: "initrd.img",
		Mode: 0600,
		Size: int64(len(initrd)),
	}
	err = tw.WriteHeader(hdr)
	if err != nil {
		return buf, err
	}
	_, err = tw.Write(initrd)
	if err != nil {
		return buf, err
	}
	err = tw.Close()
	if err != nil {
		return buf, err
	}
	return buf, nil
}

func outputImg(image, filename string, bzimage []byte, initrd []byte, args ...string) error {
	log.Debugf("output img: %s %s", image, filename)
	log.Infof("  %s", filename)
	buf, err := tarInitrdKernel(bzimage, initrd)
	if err != nil {
		return err
	}
	img, err := dockerRunInput(buf, append([]string{image}, args...)...)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, img, os.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

func outputISO(image, filename string, bzimage []byte, initrd []byte, args ...string) error {
	log.Debugf("output iso: %s %s", image, filename)
	log.Infof("  %s", filename)
	buf, err := tarInitrdKernel(bzimage, initrd)
	if err != nil {
		return err
	}
	iso, err := dockerRunInput(buf, append([]string{image}, args...)...)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, iso, os.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

func outputKernelInitrd(base string, bzimage []byte, initrd []byte, cmdline string) error {
	log.Debugf("output kernel/initrd: %s %s", base, cmdline)
	log.Infof("  %s %s %s", base+"-bzImage", base+"-initrd.img", base+"-cmdline")
	err := ioutil.WriteFile(base+"-initrd.img", initrd, os.FileMode(0644))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(base+"-bzImage", bzimage, os.FileMode(0644))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(base+"-cmdline", []byte(cmdline), os.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}
