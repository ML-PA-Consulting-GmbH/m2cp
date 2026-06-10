package cmd

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	gql "m2cpcli/graphql"
	"testing"
)

func TestLatestRevisionByVersionNumber_EmptyList(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	revisions := []gql.SnapRevision{}
	rev, err := latestRevisionByRevisionNumber(revisions)
	assert.Error(t, err)
	assert.Nil(t, rev)
	assert.Equal(t, "input list is empty", err.Error())
}

func TestLatestRevisionByVersionNumber_SingleEntry(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	revisions := []gql.SnapRevision{
		{
			Revision:    1,
			SnapVersion: "1.2.3",
		},
	}
	rev, err := latestRevisionByRevisionNumber(revisions)
	assert.NoError(t, err)
	assert.NotNil(t, rev)
	assert.Equal(t, "1.2.3", rev.SnapVersion)
}

func TestLatestRevisionByVersionNumber_DifferentPatchVersions(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	revisions := []gql.SnapRevision{
		{
			Revision:    1,
			SnapVersion: "0.3.1",
		},
		{
			Revision:    3,
			SnapVersion: "0.3.11",
		},
		{
			Revision:    2,
			SnapVersion: "0.3.2",
		},
	}

	rev, err := latestRevisionByRevisionNumber(revisions)
	assert.NoError(t, err)
	assert.NotNil(t, rev)
	assert.Equal(t, "0.3.11", rev.SnapVersion)
}
