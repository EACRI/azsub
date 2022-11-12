package azsub

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// ListBlobs at prefix returns a slice of blob names under the container prefix
func ListBlobsAtPrefix(ctx context.Context, cc *container.Client, prefix string) ([]string, error) {

	var blobsNames []string

	pager := cc.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{Snapshots: true, Versions: true},
		Prefix:  &prefix,
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return blobsNames, err
		}
		for _, blob := range resp.Segment.BlobItems {
			blobsNames = append(blobsNames, *blob.Name)
		}
	}

	return blobsNames, nil
}
