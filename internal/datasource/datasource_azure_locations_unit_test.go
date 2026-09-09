package datasource

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type testCredential struct{}

func (testCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type failingTransport struct {
	err error
}

func (t failingTransport) Do(*http.Request) (*http.Response, error) {
	return nil, t.err
}

func TestSubscriptionIDReturnsPagingError(t *testing.T) {
	expectedErr := errors.New("subscription request failed")
	options := &arm.ClientOptions{ClientOptions: policy.ClientOptions{
		Retry:     policy.RetryOptions{MaxRetries: -1},
		Transport: failingTransport{err: expectedErr},
	}}

	_, err := subscriptionId(t.Context(), testCredential{}, types.StringNull(), types.StringNull(), options)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected paging error %q, got %v", expectedErr, err)
	}
}

func TestCreateLocationMapsReturnsLocationPagingDiagnostic(t *testing.T) {
	expectedErr := errors.New("location request failed")
	options := &arm.ClientOptions{ClientOptions: policy.ClientOptions{
		Retry:     policy.RetryOptions{MaxRetries: -1},
		Transport: failingTransport{err: expectedErr},
	}}
	var diagnostics diag.Diagnostics

	subscriptionID, locations := createLocationMapsWithCredential(t.Context(), testCredential{}, types.StringValue("00000000-0000-0000-0000-000000000000"), types.StringNull(), options, &diagnostics)

	if subscriptionID == "" {
		t.Fatal("expected the configured subscription ID to be preserved")
	}
	if !diagnostics.HasError() {
		t.Fatal("expected a location paging diagnostic")
	}
	if !strings.Contains(diagnostics.Errors()[0].Detail(), expectedErr.Error()) {
		t.Fatalf("expected diagnostic to contain %q, got %q", expectedErr, diagnostics.Errors()[0].Detail())
	}
	if len(locations["locs"]) != 0 || len(locations["locs_from_display_name"]) != 0 {
		t.Fatalf("expected no locations after a paging failure, got %#v", locations)
	}
}
