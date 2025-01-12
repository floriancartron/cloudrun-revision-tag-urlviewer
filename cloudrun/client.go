package cloudrun

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
)

// Row represents a row of Cloud Run data
type Row struct {
	Date             string `json:"date"`
	Url              string `json:"url"`
	IdentifyingLabel string `json:"identifyinglabel"`
	Service          string `json:"service"`
	BaseUrl          string `json:"baseurl"`
	RevisionTag      string `json:"revisiontag"`
}

// GetCloudRunData fetches data from Cloud Run
func GetCloudRunData(project string, location string, identifyingLabel string) ([]Row, error) {
	var rows []Row
	var mu sync.Mutex // Mutex to safely append to rows
	var wg sync.WaitGroup

	// Initialize Cloud Run client
	ctx := context.Background()
	sc, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudrun services client: %v", err)
	}
	defer sc.Close()

	rc, err := run.NewRevisionsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudrun revisions client: %v", err)
	}
	defer rc.Close()

	it := sc.ListServices(ctx, &runpb.ListServicesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, location),
	})

	for {
		service, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list services: %v", err)
		}

		// Iterate over the traffic tags for the service
		for _, t := range service.GetTraffic() {
			if t.Tag != "" {
				wg.Add(1) // Increment WaitGroup counter

				// Call the helper function as a goroutine
				go fetchAndAppendRevision(ctx, rc, identifyingLabel, service, t.Revision, t.Tag, &rows, &mu, &wg)
			}
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return rows, nil
}

// fetchAndAppendRevision fetches revision details and appends them to the shared rows slice
func fetchAndAppendRevision(
	ctx context.Context,
	rc *run.RevisionsClient,
	identifyingLabel string,
	service *runpb.Service,
	revisionName string,
	tag string,
	rows *[]Row,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
) {
	defer wg.Done() // Decrement WaitGroup counter when done

	// Prepare the revision request
	req := &runpb.GetRevisionRequest{
		Name: fmt.Sprintf("%s/revisions/%s", service.Name, revisionName),
	}

	// Fetch revision details
	revision, err := rc.GetRevision(ctx, req)
	if err != nil {
		fmt.Printf("failed to get revision %s: %v\n", revisionName, err)
		return
	}

	// Convert the creation time
	createTime := time.Unix(revision.GetCreateTime().GetSeconds(), 0)

	// Construct the Row struct
	row := Row{
		Date:             createTime.Format("2006-01-02 15:04:05"),
		Url:              fmt.Sprintf("<a target=\"_blank\" href=\"https://%s.%s\">%s.%s</a>", tag, service.Annotations["baseurl"], tag, service.Annotations["baseurl"]),
		IdentifyingLabel: fmt.Sprintf("%s=%s", identifyingLabel, service.Labels[identifyingLabel]),
		Service:          getConsoleServiceUrl(service.Name),
		BaseUrl:          service.Annotations["baseurl"],
		RevisionTag:      tag,
	}

	// Append row to the shared slice
	mu.Lock()
	*rows = append(*rows, row)
	mu.Unlock()
}

func getConsoleServiceUrl(serviceFullName string) string {
	serviceNameSplitted := strings.Split(serviceFullName, "/")
	return fmt.Sprintf("<a target=\"_blank\" href=\"https://console.cloud.google.com/run/detail/%s/%s/revisions?inv=1&invt=AbmmaA&project=%s\">%s</a>", serviceNameSplitted[3], serviceNameSplitted[5], serviceNameSplitted[1], serviceNameSplitted[5])
}
