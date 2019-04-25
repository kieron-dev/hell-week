package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	flag "github.com/spf13/pflag"
	"golang.org/x/sys/unix"
)

func main() {
	rootFS := flag.String("rootfs", "", "root filesystem path")
	cgroup := flag.String("cgroup", "", "cgroup")
	isChild := flag.Bool("child", false, "execute child process")
	flag.Parse()
	args := flag.Args()

	var exitCode int

	if *isChild {
		exitCode = child(*rootFS, args)
	} else {
		exitCode = parent(*cgroup)
	}
	os.Exit(exitCode)
}

func parent(cgroup string) int {
	if cgroup != "" {
		addSelfToCgroup(cgroup)
	}
	cmd := exec.Command("/proc/self/exe", append([]string{"--child"}, os.Args[1:]...)...)
	cmd.SysProcAttr = &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
	return getExitCode(cmd)
}

func child(rootFS string, args []string) int {
	if rootFS != "" {
		pivotRoot(rootFS)
	}
	cmd := exec.Command(args[0], args[1:]...)
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

func pivotRoot(rootFS string) {
	oldDir := path.Join(rootFS, "old")
	must(unix.Mount(rootFS, rootFS, "", unix.MS_BIND, ""))
	must(unix.Mount("", rootFS, "", unix.MS_PRIVATE, ""))
	must(unix.Mount(rootFS, rootFS, "", unix.MS_BIND, ""))
	must(os.MkdirAll(oldDir, 0700))
	must(unix.PivotRoot(rootFS, oldDir))
	must(os.Chdir("/"))
}

func addSelfToCgroup(cgroup string) {
	cgroupDir := path.Join("/sys/fs/cgroup", cgroup)
	must(os.MkdirAll(cgroupDir, 0700))
	if strings.HasPrefix(cgroup, "cpuset") {
		populateCPUSetDefaults(cgroupDir)
	}
	must(ioutil.WriteFile(path.Join(cgroupDir, "tasks"), []byte(fmt.Sprintf("%d", os.Getpid())), 0644))
}

func populateCPUSetDefaults(cgroupDir string) {
	must(ioutil.WriteFile(path.Join(cgroupDir, "cpuset.cpus"), []byte("0"), 0644))
	must(ioutil.WriteFile(path.Join(cgroupDir, "cpuset.mems"), []byte("0"), 0644))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
