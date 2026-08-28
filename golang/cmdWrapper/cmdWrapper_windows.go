//go:build windows

package cmdWrapper

import (
	"os/exec"
	"syscall"
)

func applySysProcAttr(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return
	}
	cmd.SysProcAttr.HideWindow = true
}

func applySysProcAttrVisible(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		return
	}
	cmd.SysProcAttr.HideWindow = false
}
