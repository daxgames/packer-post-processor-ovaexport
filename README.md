'ovaexport' a Packer Post Processor
===================================

This Packer post processor uses the VMware ovftool binaries to export generated
VMware vmx files to ova/ovf formats.

Requirements
============

1. [Packer](https://www.packer.io/downloads.html) 0.7.5
1. Refer to [Developing Packer](https://github.com/mitchellh/packer#developing-packer) to install the software required to compile this code.
  * If you downloaded and installed the Packer binaries you can stop after installing Gox.
1. [VMware OVFTool](https://my.vmware.com/web/vmware/details?productId=491&downloadGroup=OVFTOOL410) insalled and in your search path.

Installation
============

1. Run the following:

```
$ go get github.com/daxgames/packer-post-procesror-ovaexport
$ go install github.com/daxgames/packer-post-processor-ovaexport
```

Usage
=====

Export type is determined by the target filename extension specified in the
Packer template file.  As shown below with default values:

```
  {
    "type": "ovaexport",
    "target": "ova/vmware/rhel-6.6-chef12.0.3.ova",
    "disk_mode": "thick",
    "remove_floppy": "false",
    "remove_optical_drive": "false",
    "remove_ethernet": "false",
    "compress": 9,
    "only": ["vmware-iso"]
  }
```

Thanks and Credit
=================

Thanks to Mitchell Hashimoto, and others, for [Packer](https://github.com/mitchellh/packer#developing-packer) and the vsphere post
processor on which this code is based.

Thanks to Ian McCracken who wrote [packer-post-processor-ovftool](https://github.com/iancmcc/packer-post-processor-ovftool).  I could not
get it to work with the latest Packer 0.7.5 but it inspired me to cobble this
together instead.

