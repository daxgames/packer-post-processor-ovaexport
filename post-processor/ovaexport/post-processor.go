package ovaexport

import (
  "bytes"
  "fmt"
  "log"
  "os/exec"
  "strings"
  vmwarecommon "github.com/mitchellh/packer/builder/vmware/common"
  "github.com/mitchellh/packer/common"
  "github.com/mitchellh/packer/helper/config"
  "github.com/mitchellh/packer/packer"
  "github.com/mitchellh/packer/template/interpolate"
)

var builtins = map[string]string{
  "mitchellh.vmware": "vmware",
}

type Config struct {
  common.PackerConfig `mapstructure:",squash"`

  DiskMode              string `mapstructure:"disk_mode"`
  Target                string `mapstructure:"target"`
  RemoveEthernet        string `mapstructure:"remove_ethernet"`
  RemoveFloppy          string `mapstructure:"remove_floppy"`
  RemoveOpticalDrive    string `mapstructure:"remove_optical_drive"`
  Compression           uint   `mapstructure:"compression"`

  ctx interpolate.Context
}

type PostProcessor struct {
  config Config
}

func (p *PostProcessor) RemoveFloppy(vmx string, ui packer.Ui, artifact packer.Artifact ) error {
  ui.Message(fmt.Sprintf("Removing floppy from %s", vmx))
  vmxData, err := vmwarecommon.ReadVMX(vmx)
  if err != nil {
    return err
  }
  for k, _ := range vmxData {
    if strings.HasPrefix(k, "floppy0.") {
      delete(vmxData, k)
    }
  }
  vmxData["floppy0.present"] = "FALSE"
  if err := vmwarecommon.WriteVMX(vmx, vmxData); err != nil {
    return err
  }
  return nil
}

func (p *PostProcessor) RemoveEthernet(vmx string, ui packer.Ui, artifact packer.Artifact) error {
  ui.Message(fmt.Sprintf("Removing ethernet intercace from %s", vmx))
  vmxData, err := vmwarecommon.ReadVMX(vmx)
  if err != nil {
    return err
  }

  for k, _ := range vmxData {
    if strings.HasPrefix(k, "ethernet0.") {
      delete(vmxData, k)
    }
  }

  vmxData["ethernet0.present"] = "FALSE"
  if err := vmwarecommon.WriteVMX(vmx, vmxData); err != nil {
    return err
  }

  return nil
}

func (p *PostProcessor) RemoveOpticalDrive(vmx string, ui packer.Ui, artifact packer.Artifact) error {
  ui.Message(fmt.Sprintf("Removing optical drive from %s", vmx))
  vmxData, err := vmwarecommon.ReadVMX(vmx)
  if err != nil {
    return err
  }

  for k, _ := range vmxData {
    if strings.HasPrefix(k, "ide1:0.file") {
      delete(vmxData, k)
    }
  }

  vmxData["ide1:0.present"] = "FALSE"

  if err := vmwarecommon.WriteVMX(vmx, vmxData); err != nil {
    return err
  }
  return nil
}


func (p *PostProcessor) Configure(raws ...interface{}) error {
  err := config.Decode(&p.config, &config.DecodeOpts{
    Interpolate: true,
    InterpolateFilter: &interpolate.RenderFilter{
      Exclude: []string{},
    },
  }, raws...)
  if err != nil {
    return err
  }

  // Defaults
  if p.config.DiskMode == "" {
    p.config.DiskMode = "thick"
  }

  if p.config.RemoveEthernet == "" {
    p.config.RemoveEthernet = "false"
  }

  if p.config.RemoveFloppy == "" {
    p.config.RemoveFloppy = "false"
  }

  if p.config.RemoveOpticalDrive == "" {
    p.config.RemoveOpticalDrive = "false"
  }

  if ! (p.config.Compression > 0) {
    p.config.Compression = 9
  }

  // Accumulate any errors
  errs := new(packer.MultiError)

  if !(p.config.Compression >= 0 && p.config.Compression <= 9) {
    errs = packer.MultiErrorAppend(
    errs, fmt.Errorf("Invalid compression level. Must be between 1 and 9, or 0 for no compression."))
  }

  if _, err := exec.LookPath("ovftool"); err != nil {
    errs = packer.MultiErrorAppend(
      errs, fmt.Errorf("ovftool not found: %s", err))
  }

  // First define all our templatable parameters that are _required_
  var compression = string(p.config.Compression)
  templates := map[string]*string{
    "disk_mode":            &p.config.DiskMode,
    "target":               &p.config.Target,
    "remove_ethernet":      &p.config.RemoveEthernet,
    "remove_floppy":        &p.config.RemoveFloppy,
    "remove_optical_drive": &p.config.RemoveOpticalDrive,
    "compression":          &compression,
  }

  for key, ptr := range templates {
    if *ptr == "" {
      errs = packer.MultiErrorAppend(
        errs, fmt.Errorf("%s must be set", key))
    }
  }

  if len(errs.Errors) > 0 {
    return errs
  }

  return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
  if _, ok := builtins[artifact.BuilderId()]; !ok {
    return nil, false, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
  }

  vmx := ""
  for _, path := range artifact.Files() {
    if strings.HasSuffix(path, ".vmx") {
      vmx = path
      break
    }
  }

  if vmx == "" {
    return nil, false, fmt.Errorf("VMX file not found")
  }

  if p.config.RemoveEthernet == "true" {
    if err := p.RemoveEthernet(vmx, ui, artifact); err != nil {
      return nil, false, fmt.Errorf("Removing ethernet interfaces from VMX failed!")
    }
  }

  if p.config.RemoveFloppy == "true" {
    if err := p.RemoveFloppy(vmx, ui, artifact); err != nil {
      return nil, false, fmt.Errorf("Removing floppy drive from VMX failed!")
    }
  }

  if p.config.RemoveOpticalDrive == "true" {
    if err := p.RemoveOpticalDrive(vmx, ui, artifact); err != nil {
      return nil, false, fmt.Errorf("Removing CD/DVD Drive from VMX failed!")
    }
  }

  args := []string{
    "--acceptAllEulas",
    fmt.Sprintf("--diskMode=%s", p.config.DiskMode),
    fmt.Sprintf("--compress=%s", string(p.config.Compression)),
    fmt.Sprintf("%s", vmx),
    fmt.Sprintf("%s", p.config.Target),
  }

  ui.Message(fmt.Sprintf("Exporting %s to %s", vmx, p.config.Target))
  var out bytes.Buffer
  log.Printf("Starting ovftool with parameters: %s", strings.Join(args, " "))
  cmd := exec.Command("ovftool", args...)
  cmd.Stdout = &out
  if err := cmd.Run(); err != nil {
    return nil, false, fmt.Errorf("Failed: %s\nStdout: %s", err, out.String())
  }

  ui.Message(fmt.Sprintf("%s", out.String()))

  return artifact, false, nil
}
