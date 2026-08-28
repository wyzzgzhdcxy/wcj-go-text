//go:build !windows

package cmdWrapper

import "os/exec"

func applySysProcAttr(cmd *exec.Cmd) {}

func applySysProcAttrVisible(cmd *exec.Cmd) {}
