package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

func main() {
	cgroup := flag.String("cgroup", "", "cgroup path (should start with memory/)")
	max := flag.Int("max", 0, "maximum allowed memory in bytes")
	flag.Parse()

	if *cgroup == "" || *max == 0 {
		panic("eh?")
	}

	setMemLimit(*cgroup, *max)
}

func setMemLimit(cgroup string, memBytes int) {
	base := path.Join("/sys/fs/cgroup", cgroup)
	if _, err := os.Stat(base); os.IsNotExist(err) {
		panic(err)
	}
	must(ioutil.WriteFile(path.Join(base, "memory.limit_in_bytes"), []byte(fmt.Sprintf("%d", memBytes)), 0644))
	// memsw is mem+swap, so setting to mem limit means above line is redundant?
	must(ioutil.WriteFile(path.Join(base, "memory.memsw.limit_in_bytes"), []byte(fmt.Sprintf("%d", memBytes)), 0644))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
