//go:build !windows

package crypto

import (
	"github.com/google/go-tpm/legacy/tpm2"
	"io"
)

func openTpm() (io.ReadWriteCloser, error) {
	return tpm2.OpenTPM(pathTpmDevice)
}
