package tests

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func Test_MobileApiBandwidthAbuse(t *testing.T) {
	someId := uuid.New()

	noOfRequest := 130
	noOfWorkers := 20
	responseCh := make(chan int, noOfRequest)

	eg := errgroup.Group{}
	eg.SetLimit(noOfWorkers)
	for i := 0; i < noOfRequest; i++ {
		eg.Go(func() error {
			resp := e2e_tests.DoAPIGet(t, "/mobile/devices/"+someId.String()+"/browser_extensions", nil)

			responseCh <- resp.StatusCode

			return nil
		})
	}
	require.NoError(t, eg.Wait())
	close(responseCh)

	var got404, got429 int
	for code := range responseCh {
		switch code {
		case http.StatusNotFound:
			got404++
		case http.StatusTooManyRequests:
			got429++
		default:
			t.Fatalf("Unexpected code: %v", code)
		}
	}
	// Default rate limit is 100 per minute.
	// So we expect around 100 - 404, and around 30 - 429
	require.InDelta(t, 100, got404, 2.0)
	require.InDelta(t, 30, got429, 2.0)
}

func Test_BrowserExtensionApiBandwidthAbuse(t *testing.T) {
	someId := uuid.New()

	noOfRequest := 130
	noOfWorkers := 20
	responseCh := make(chan int, noOfRequest)

	eg := errgroup.Group{}
	eg.SetLimit(noOfWorkers)
	for i := 0; i < noOfRequest; i++ {
		eg.Go(func() error {
			resp := e2e_tests.DoAPIGet(t, "/browser_extensions/"+someId.String(), nil)

			responseCh <- resp.StatusCode

			return nil
		})
	}
	require.NoError(t, eg.Wait())
	close(responseCh)

	var got404, got429 int
	for code := range responseCh {
		switch code {
		case http.StatusNotFound:
			got404++
		case http.StatusTooManyRequests:
			got429++
		default:
			t.Fatalf("Unexpected code: %v", code)
		}
	}
	// Default rate limit is 100 per minute.
	// So we expect around 100 - 404, and around 30 - 429
	require.InDelta(t, 100, got404, 2.0)
	require.InDelta(t, 30, got429, 2.0)
}
