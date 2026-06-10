package tools

import (
	"os"
	"strings"
)

func IsRunningInAzurePipeline() bool {
	return strings.ToLower(os.Getenv("TF_BUILD")) == "true"
}
