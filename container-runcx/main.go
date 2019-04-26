package main

import (
	"os"
	"os/exec"

	flag "github.com/spf13/pflag"
)

func main() {
	id := flag.String("id", "", "container ID (required)")
	flag.Parse()
	args := flag.Args()

	if *id == "" {
		panic("--id required")
	}

	cmd := exec.Command("runc", append([]string{"exec", *id}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

}
