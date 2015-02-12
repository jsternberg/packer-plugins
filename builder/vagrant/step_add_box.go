package vagrant

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepAddBox struct {
	BoxUrl string

	tempDir string
}

func (s *StepAddBox) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	tempDir, err := ioutil.TempDir(os.TempDir(), "packer-vagrant")
	if err != nil {
		ui.Error(fmt.Sprintf("Could not create a temporary directory: %s", err))
		return multistep.ActionHalt
	}
	s.tempDir = tempDir

	ui.Say(fmt.Sprintf("Created temporary directory: %s", s.tempDir))

	ui.Say(fmt.Sprintf("Downloading box from %s", s.BoxUrl))
	path := filepath.Join(s.tempDir, "virtualbox.box")
	if err = s.downloadFile(path); err != nil {
		ui.Say(fmt.Sprintf("Could not create file: %s: %s", path, err))
		return multistep.ActionHalt
	}

	ui.Say("Extracting the box data")
	if err = s.extractBox(path); err != nil {
		ui.Say(fmt.Sprintf("Could not extract the archive: %s", err))
		return multistep.ActionHalt
	}

	ovf := filepath.Join(s.tempDir, "box.ovf")
	if _, err = os.Stat(ovf); err != nil {
		ui.Say("Could not find box.ovf file in the box archive")
		return multistep.ActionHalt
	}
	state.Put("ovf", ovf)
	return multistep.ActionContinue
}

func (s *StepAddBox) downloadFile(path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(s.BoxUrl)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *StepAddBox) extractBox(path string) error {
	dir := filepath.Dir(path)
	cmd := exec.Command("tar", "-C", dir, "-xzf", path)
	return cmd.Run()
}

func (s *StepAddBox) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if s.tempDir != "" {
		ui.Say(fmt.Sprintf("Removing temporary directory: %s", s.tempDir))
		os.RemoveAll(s.tempDir)
	}
}
