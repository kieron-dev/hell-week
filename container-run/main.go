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
	volume := flag.String("volume", "", "<local-path>:<container-path>")
	isChild := flag.Bool("child", false, "execute child process")
	flag.Parse()
	args := flag.Args()

	var exitCode int

	if *isChild {
		exitCode = child(*rootFS, *volume, args)
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
		Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWNS | unix.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
	return getExitCode(cmd)
}

func child(rootFS string, volume string, args []string) int {
	oldDir := "/"
	if rootFS != "" {
		oldDir = pivotRoot(rootFS)
		mountProc()
	}
	if volume != "" {
		mountVolume(volume, oldDir)
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

func pivotRoot(rootFS string) string {
	old := "/old"
	oldDir := path.Join(rootFS, old)
	must(unix.Mount(rootFS, rootFS, "", unix.MS_BIND, ""))
	must(unix.Mount("", rootFS, "", unix.MS_PRIVATE, ""))
	must(unix.Mount(rootFS, rootFS, "", unix.MS_BIND, ""))
	must(os.MkdirAll(oldDir, 0700))
	must(unix.PivotRoot(rootFS, oldDir))
	must(os.Chdir("/"))
	return old
}

func mountProc() {
	must(os.MkdirAll("/proc", 0755))
	must(unix.Mount("proc", "/proc", "proc", 0, ""))
}

func mountVolume(volume, oldDir string) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		panic("can't parse volume param")
	}
	source, target := parts[0], parts[1]
	source = path.Join(oldDir, source)
	mustExist(source)
	must(os.MkdirAll(target, 0755))
	must(unix.Mount(source, target, "", unix.MS_BIND, ""))
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

func mustExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err != nil {
			panic(err)
		}
	}
}
