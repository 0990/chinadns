package version

import (
	"fmt"
	"runtime"
)

const Name = "chinadns"

var (
	Version   string
	GitCommit string
)

func String() string {
	return fmt.Sprintf("%s-%s", Name, Version)
}

func BuildString() string {
	return fmt.Sprintf("%s/%s, %s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}
