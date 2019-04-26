package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pborman/uuid"
	flag "github.com/spf13/pflag"
)

func main() {

	rootFS := flag.String("rootfs", "", "root filesystem path")
	cgroup := flag.String("cgroup", "", "cgroup")
	volume := flag.String("volume", "", "<local-path>:<container-path>")
	flag.Parse()
	args := flag.Args()

	if *rootFS == "" {
		panic("you need to specify --rootfs")
	}

	if len(args) == 0 {
		panic("specify something to run")
	}

	spec := buildSpec(*rootFS, *cgroup, *volume, args)

	bundle, err := ioutil.TempDir("/tmp", "bundle-")
	if err != nil {
		panic(err)
	}

	data, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(bundle, "config.json"), data, 0644)
	if err != nil {
		panic(err)
	}

	id := uuid.New()[:8]

	cmd := exec.Command("runc", "run", "--bundle", bundle, id)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func buildSpec(rootFS, cgroup, volume string, args []string) *specs.Spec {
	spec := specconv.Example()
	spec.Process.Terminal = false
	spec.Process.Args = args
	spec.Root.Path = rootFS
	spec.Root.Readonly = false

	if cgroup != "" {
		spec.Linux.CgroupsPath = cgroup
	}

	if volume != "" {
		parts := strings.Split(volume, ":")
		if len(parts) != 2 {
			panic("can't understand volume format")
		}
		spec.Mounts = append(spec.Mounts, specs.Mount{
			Source:      parts[0],
			Destination: parts[1],
			Options:     []string{"bind"},
		})
	}
	return spec
}
