package legacy

import (
	"context"
	"fmt"
)

func GetSnapRevisionByNameArchRevision(ctx context.Context, snapName, snapArchitecture string, revision int) (*GetSnapRevisionsByNameArchRevisionSnapRevisionsSnapRevisionCollectionSegmentItemsSnapRevision, error) {
	snapRevisions, err := GetSnapRevisionsByNameArchRevision(ctx, snapName, snapArchitecture, revision)
	if err != nil {
		return nil, err
	}
	if len(snapRevisions.SnapRevisions.Items) != 1 {
		return nil, fmt.Errorf("expected 1 snap revision, got %d", len(snapRevisions.GetSnapRevisions().Items))
	}

	return snapRevisions.GetSnapRevisions().Items[0], nil
}
