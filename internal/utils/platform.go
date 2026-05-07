package utils

import (
	"fmt"
	"runtime"
)

type Platform struct {
	OS   string
	Arch string
}

func DetectPlatform() (Platform, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "linux", "darwin":
		break
	default:
		return Platform{}, fmt.Errorf("unsupported OS: %s (only linux and darwin are supported)", os)
	}

	switch arch {
	case "amd64", "arm64":
		break
	default:
		return Platform{}, fmt.Errorf("unsupported architecture: %s (only amd64 and arm64 are supported)", arch)
	}

	return Platform{OS: os, Arch: arch}, nil
}

func (p Platform) String() string {
	return fmt.Sprintf("%s-%s", p.OS, p.Arch)
}

func (p Platform) BinaryName() string {
	return "orbit-" + p.String()
}

func (p Platform) ArchiveName() string {
	return p.BinaryName() + ".tar.gz"
}
