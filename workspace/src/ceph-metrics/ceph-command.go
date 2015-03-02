package main

import (
	"bytes"
	"io"

	"os/exec"
)

func cephCommand(args ...string) (io.Reader, error) {
	args = append(args, "-f", "json")
	out, err := exec.Command("/usr/bin/ceph", args...).Output()
	return bytes.NewReader(out), err
}
