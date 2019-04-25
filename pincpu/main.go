package main

import (
	"io/ioutil"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

func main() {
	cgroup := flag.String("cgroup", "", "cgroup path (should start with cpuset/)")
	cpus := flag.String("cpus", "", "CPU(s) to use, e.g. '0-3', or '1,3'")
	flag.Parse()

	if *cgroup == "" || *cpus == "" {
		panic("eh?")
	}

	setCPUs(*cgroup, *cpus)
}

func setCPUs(cgroup, cpus string) {
	base := path.Join("/sys/fs/cgroup", cgroup)
	if _, err := os.Stat(base); os.IsNotExist(err) {
		panic(err)
	}
	err := ioutil.WriteFile(path.Join(base, "cpuset.cpus"), []byte(cpus), 0644)
	if err != nil {
		panic(err)
	}
}
