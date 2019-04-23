package main

import (
	"os"
	"os/exec"
	"syscall"
)

func main() {
	var exitCode int

	switch os.Args[1] {
	case "run":
		exitCode = parent()
	case "child":
		exitCode = child()
	default:
		panic("wat should I do")
	}
	os.Exit(exitCode)
}

func parent() int {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
	return getExitCode(cmd)
}

func child() int {
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
	return getExitCode(cmd)
}

func getExitCode(cmd *exec.Cmd) int {
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	var exitCode int
	if status.Signaled() {
		exitCode = 128 + int(status.Signal())
	} else {
		exitCode = status.ExitStatus()
	}
	return exitCode
}
