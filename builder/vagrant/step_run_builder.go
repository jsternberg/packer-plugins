package vagrant

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/virtualbox/ovf"
	"github.com/mitchellh/packer/packer"
)

type StepRunBuilder struct {
	builder *ovf.Builder
}

func (s *StepRunBuilder) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	hook := state.Get("hook").(packer.Hook)
	cache := state.Get("cache").(packer.Cache)

	user, err := user.Current()
	if err != nil {
		ui.Error(fmt.Sprintf("Could not find the user's home directory: %s", err))
		return multistep.ActionHalt
	}
	ssh_key_path := filepath.Join(user.HomeDir, ".vagrant.d/insecure_private_key")

	if _, err := os.Stat(ssh_key_path); err != nil {
		ui.Error("The insecure private key could not be found")
		return multistep.ActionHalt
	}

	pc := state.Get("config").(*Config).PackerConfig
	config := map[string]interface{}{
		"packer_build_name":     pc.PackerBuildName,
		"packer_builder_type":   pc.PackerBuilderType,
		"packer_debug":          pc.PackerDebug,
		"packer_force":          pc.PackerForce,
		"packer_user_variables": pc.PackerUserVars,
		"source_path":           state.Get("ovf").(string),
		"ssh_key_path":          ssh_key_path,
		"ssh_username":          "vagrant",
		"shutdown_command":      "sudo shutdown -h now",
		"headless":              true,
	}

	builder := new(ovf.Builder)
	if warnings, err := builder.Prepare(config); err != nil {
		for _, warning := range warnings {
			ui.Error(warning)
		}
		ui.Error("Failed to prepare the virtualbox-ovf builder")
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.builder = builder

	artifact, err := builder.Run(ui, hook, cache)
	if err != nil {
		return multistep.ActionHalt
	}
	state.Put("artifact", artifact)
	return multistep.ActionContinue
}

func (s *StepRunBuilder) Cleanup(state multistep.StateBag) {
	if s.builder != nil {
		s.builder.Cancel()
	}
}
