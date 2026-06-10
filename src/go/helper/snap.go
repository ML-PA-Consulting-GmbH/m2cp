package helper

import (
	"bytes"
	"fmt"
	"io"
	"m2cpcli/graphql"
	"m2cpcli/tools"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/appendblob"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// FindRequestedRevision finds the requested revision in the list of revisions. If the requested revision is "latest", the
// revision with the highest revision number is returned. If the requested revision is not found, an error is returned.
func FindRequestedRevision(requestedRevision string, revisions []graphql.SnapRevision) (*graphql.SnapRevision, error) {
	if len(revisions) == 0 {
		return nil, fmt.Errorf("no revisions")
	}
	if requestedRevision == "latest" {
		// Find snap-revision with the highest revision number
		revision := &revisions[0]
		for _, snapRevision := range revisions {
			if snapRevision.Revision > revision.Revision {
				revision = &snapRevision
			}
		}
		return revision, nil
	} else {
		// Find snap-revision with requested revision number
		var revisionToFind int32
		intRevisionToFind, err := strconv.Atoi(requestedRevision)
		if err != nil {
			return nil, err
		}
		revisionToFind = int32(intRevisionToFind)

		for _, revision := range revisions {
			if revision.Revision == revisionToFind {
				return &revision, nil
			}
		}
	}
	return nil, fmt.Errorf("revision %s not found", requestedRevision)
}

// SnapByNameAndArchitecture returns the snap declaration (with revision and declaration) for the given snap name and architecture.
// If the snap is not found, an error is returned.
func SnapByNameAndArchitecture(ctx context.Context, snapName string, snapDeviceArchitecture string) (*graphql.SnapDeclaration, error) {
	dbId, err := graphql.SnapDatabaseIdByNameAndArchitecture(ctx, snapName, snapDeviceArchitecture)
	if err != nil {
		return nil, err
	}
	return graphql.SnapDeclarationByDatabaseId(ctx, dbId, true, true)
}

func RetrieveSnapFilenameFromUrl(url string) (string, error) {
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	r, err := regexp.Compile(`([a-zA-Z0-9_.+-]+\.snap)`)
	if err != nil {
		return "", err
	}
	matches := r.FindStringSubmatch(filename)
	if matches == nil {
		return "", fmt.Errorf("failed to retrieve snap filename from URL \"%s\"", url)
	}
	if len(matches) > 2 {
		return "", fmt.Errorf("multiple snap names found in URL \"%s\"", url)
	}

	return matches[1], nil
}

// CreateFallbackOutputPath tries to make up a filename from the given command line arguments by best effort.
func CreateFallbackOutputPath(cmd *cobra.Command, args []string) string {
	var revision string
	if cmd.Flags().Changed("revision") {
		revision = cmd.Flag("revision").Value.String()
	} else {
		revision = "latest"
	}

	var result string
	switch len(args) {
	case 1:
		// only appId is expected
		snapId := args[0]
		result = fmt.Sprintf("mlpa_%s_%s.snap", snapId, revision)
	case 2:
		// snapName and snapDeviceArchitecture are expected
		snapName := args[0]
		snapDeviceArchitecture := args[1]
		result = fmt.Sprintf("mlpa_%s_%s_%s.snap", snapName, revision, snapDeviceArchitecture)
	}

	return result
}

// DownloadSnap downloads a snap from the given URL and saves it to the given output path. If withProgress is true, a progress bar is printed to the console.
// The output path is returned if the download was successful.
func DownloadSnap(url string, givenOutputPath string, fallbackOutputPath string, withProgress bool) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Determine binary filename by URL if no output path is given
	// Azure blob storage does not give the name by the Content-Disposition header :(
	var outputPath = ""
	if givenOutputPath == "" {
		outputPath, err = RetrieveSnapFilenameFromUrl(url)
		if err != nil {
			if fallbackOutputPath != "" {
				return fallbackOutputPath, nil
			}
			return "", err
		}
	} else {
		outputPath = givenOutputPath
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if withProgress {
		total, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		bar := pb.StartNew(total)
		reader := bar.NewProxyReader(resp.Body)
		_, err = io.Copy(out, reader)
		bar.Finish()
	} else {
		_, err = io.Copy(out, resp.Body)
	}

	if err != nil {
		return "", err
	}

	return outputPath, nil
}

func UploadToAzureBlobStorage(ctx context.Context, binary []byte, uploadUrl string) error {
	var maxChunkSize = 4 * 1024 * 1024
	chunks := tools.SplitIntoEqualChunks(binary, maxChunkSize)

	appendBlobClient, err := appendblob.NewClient(uploadUrl, nil, nil)
	if err != nil {
		return err
	}
	_, err = appendBlobClient.Create(ctx, nil)
	if err != nil {
		return err
	}
	for _, chunk := range chunks {
		_, err = appendBlobClient.AppendBlock(ctx, streaming.NopCloser(bytes.NewReader(chunk)), nil)
		if err != nil {
			return err
		}
	}
	return nil
}
