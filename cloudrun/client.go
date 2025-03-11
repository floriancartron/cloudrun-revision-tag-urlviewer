package cloudrun

import (
	"context"
	"fmt"
	"html"
	"os"
	"strings"
	"sync"
	"time"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	neg "github.com/floriancartron/cloudrun-revision-tag-urlviewer/neg"
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
func GetCloudRunData(project string, location string, identifyingLabel string, maxRevisions int) ([]Row, error) {
	var rows []Row
	var mu sync.Mutex // Mutex to safely append to rows
	var wg sync.WaitGroup

	// Initialize Cloud Run client
	ctx := context.Background()
	negs, err := neg.GetServerlessNegsUrlMasks(ctx, project, location)
	if err != nil {
		negs = map[string]string{}
		fmt.Printf("failed to get serverless neg: %v\n", err)
	}
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
		if service.GetAnnotations()["baseurl"] == "" && service.GetAnnotations()["serverless-neg"] == "" {
			fmt.Printf("service %s has no baseurl or serverless-neg annotation, it is ignored\n", service.Name)
			continue
		}
		revisionsCount := 0
		tags := service.GetTraffic() // Get the revisions tags

		// Iterate over the tags in reverse order
		// this allow to get the most recent revisions tags first
		for i := len(tags) - 1; i >= 0; i-- {
			t := tags[i]
			if t.Tag != "" {
				wg.Add(1) // Increment WaitGroup counter

				// Call the helper function as a goroutine
				go fetchAndAppendRevision(ctx, negs, rc, identifyingLabel, service, t.Revision, t.Tag, &rows, &mu, &wg)

				revisionsCount++
				if revisionsCount >= maxRevisions {
					break
				}
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
	negs map[string]string,
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
	location, err := time.LoadLocation(os.Getenv("CRRTUV_TIMEZONE"))
	if err == nil {
		createTime = createTime.In(location)
	} else {
		fmt.Println("Error loading location:", err)
	}

	baseUrl, generatedUrl, err := getRevisionTagUrl(tag, service, negs)
	if err != nil {
		fmt.Printf("%v,%s\n", err, revisionName)
		return
	}
	// Construct the Row struct
	row := Row{
		Date:             createTime.Format("2006-01-02 15:04:05"),
		Url:              generatedUrl,
		IdentifyingLabel: fmt.Sprintf("%s=%s", identifyingLabel, service.Labels[identifyingLabel]),
		Service:          getConsoleServiceUrl(service.Name),
		BaseUrl:          baseUrl,
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

func getRevisionTagUrl(tag string, service *runpb.Service, negsUrlMasks map[string]string) (string, string, error) {
	if service.Annotations["serverless-neg"] != "" {
		if urlMask, ok := negsUrlMasks[service.Annotations["serverless-neg"]]; ok {
			url := strings.Replace(strings.Replace(urlMask, "<tag>", tag, 1), "<service>", strings.Split(service.Name, "/")[5], 1)
			return html.EscapeString(urlMask), fmt.Sprintf("<a target=\"_blank\" href=\"https://%s\">%s</a>", url, url), nil
		}
	}
	if service.Annotations["baseurl"] != "" {
		return service.Annotations["baseurl"], fmt.Sprintf("<a target=\"_blank\" href=\"https://%s.%s\">%s.%s</a>", tag, service.Annotations["baseurl"], tag, service.Annotations["baseurl"]), nil
	}
	return "", "", fmt.Errorf("failed to get revision tag url for service %s", service.Name)
}
