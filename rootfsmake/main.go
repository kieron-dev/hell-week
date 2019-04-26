package main

import (
	"fmt"
	"io/ioutil"
	"os"

	flag "github.com/spf13/pflag"
	"golang.org/x/sys/unix"
)

func main() {
	image := flag.String("image", "", "path to base image")
	graph := flag.String("graph", "", "path to an empty, temporary directory")
	flag.Parse()

	if *image == "" || *graph == "" {
		panic("eh?")
	}
	mustExist(*image)
	mustExist(*graph)

	mergedDir := makeLayeredRootFS(*image, *graph)
	fmt.Print(mergedDir)
}

func makeLayeredRootFS(image, graph string) string {
	location, err := ioutil.TempDir("/tmp", "rootfs")
	if err != nil {
		panic(err)
	}
	workDir, err := ioutil.TempDir("/tmp", "work")
	if err != nil {
		panic(err)
	}

	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", image, graph, workDir)
	must(unix.Mount("overlay", location, "overlay", 0, data))
	os.Chown(location, 0777, -1)

	return location
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic(err)
	}
}
