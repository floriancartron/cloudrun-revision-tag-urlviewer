package neg

import (
	"context"
	"fmt"
	"log/slog"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

func GetServerlessNegsUrlMasks(logger *slog.Logger, ctx context.Context, project string, location string) (map[string]string, error) {
	nc, err := compute.NewRegionNetworkEndpointGroupsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create network endpoint group client: %v", err)
	}
	defer nc.Close()

	masksList := make(map[string]string)

	req := &computepb.ListRegionNetworkEndpointGroupsRequest{
		Region:  location,
		Project: project,
	}
	it := nc.List(ctx, req)
	for {
		neg, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get url masks: %v\n", err))
			return nil, err
		}
		masksList[neg.GetName()] = neg.GetCloudRun().GetUrlMask()
	}
	return masksList, nil
}
