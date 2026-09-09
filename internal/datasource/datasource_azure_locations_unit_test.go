package datasource

import (
	"context"
	"errors"
	"io"
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

type responseTransport struct {
	body string
}

func (t responseTransport) Do(request *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(t.body)),
		Request:    request,
	}, nil
}

func testClientOptions(transport policy.Transporter) *arm.ClientOptions {
	return &arm.ClientOptions{ClientOptions: policy.ClientOptions{
		Retry:     policy.RetryOptions{MaxRetries: -1},
		Transport: transport,
	}}
}

func testDependencies(transport policy.Transporter, activeSubscriptionID func(context.Context) (string, error)) azureLocationsDependencies {
	return azureLocationsDependencies{
		credential:           testCredential{},
		clientOptions:        testClientOptions(transport),
		activeSubscriptionID: activeSubscriptionID,
	}
}

func activeAzureCLISubscription(subscriptionID string) func(context.Context) (string, error) {
	return func(context.Context) (string, error) {
		return subscriptionID, nil
	}
}

func TestSubscriptionIDReturnsPagingError(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "")
	expectedErr := errors.New("subscription request failed")
	options := testClientOptions(failingTransport{err: expectedErr})

	_, err := subscriptionId(t.Context(), testDependencies(options.Transport, activeAzureCLISubscription("different-subscription")), types.StringNull(), types.StringNull())

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected paging error %q, got %v", expectedErr, err)
	}
}

func TestSubscriptionIDUsesExplicitConfiguration(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "environment-subscription")
	options := testClientOptions(failingTransport{err: errors.New("subscription request should not be made")})

	actual, err := subscriptionId(t.Context(), testDependencies(options.Transport, activeAzureCLISubscription("different-subscription")), types.StringValue("configured-subscription"), types.StringValue("configured-name"))

	if err != nil {
		t.Fatalf("expected configured subscription ID, got error: %v", err)
	}
	if actual != "configured-subscription" {
		t.Fatalf("expected configured subscription ID, got %q", actual)
	}
}

func TestSubscriptionIDUsesEnvironment(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", " environment-subscription ")
	options := testClientOptions(failingTransport{err: errors.New("subscription request should not be made")})

	actual, err := subscriptionId(t.Context(), testDependencies(options.Transport, activeAzureCLISubscription("different-subscription")), types.StringNull(), types.StringValue("configured-name"))

	if err != nil {
		t.Fatalf("expected environment subscription ID, got error: %v", err)
	}
	if actual != "environment-subscription" {
		t.Fatalf("expected trimmed environment subscription ID, got %q", actual)
	}
}

func TestSubscriptionIDUsesConfiguredDisplayName(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "")
	options := testClientOptions(responseTransport{body: `{
		"value": [
			{"subscriptionId": "first-subscription", "displayName": "First"},
			{"subscriptionId": "selected-subscription", "displayName": "Selected"}
		]
	}`})

	actual, err := subscriptionId(t.Context(), testDependencies(options.Transport, activeAzureCLISubscription("different-subscription")), types.StringNull(), types.StringValue("Selected"))

	if err != nil {
		t.Fatalf("expected display-name subscription ID, got error: %v", err)
	}
	if actual != "selected-subscription" {
		t.Fatalf("expected display-name subscription ID, got %q", actual)
	}
}

func TestSubscriptionIDUsesOnlyVisibleSubscription(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "")
	options := testClientOptions(responseTransport{body: `{
		"value": [{"subscriptionId": "only-subscription", "displayName": "Only"}]
	}`})
	activeSubscriptionCalled := false

	dependencies := testDependencies(options.Transport, func(context.Context) (string, error) {
		activeSubscriptionCalled = true
		return "different-subscription", nil
	})
	actual, err := subscriptionId(t.Context(), dependencies, types.StringNull(), types.StringNull())

	if err != nil {
		t.Fatalf("expected only visible subscription ID, got error: %v", err)
	}
	if actual != "only-subscription" {
		t.Fatalf("expected only visible subscription ID, got %q", actual)
	}
	if activeSubscriptionCalled {
		t.Fatal("expected active subscription lookup to be skipped for one visible subscription")
	}
}

func TestSubscriptionIDUsesActiveAzureCLISubscription(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "")
	options := testClientOptions(responseTransport{body: `{
		"value": [
			{"subscriptionId": "first-subscription", "displayName": "First"},
			{"subscriptionId": "selected-subscription", "displayName": "Selected"}
		]
	}`})

	actual, err := subscriptionId(t.Context(), testDependencies(options.Transport, activeAzureCLISubscription("selected-subscription")), types.StringNull(), types.StringNull())

	if err != nil {
		t.Fatalf("expected active Azure CLI subscription ID, got error: %v", err)
	}
	if actual != "selected-subscription" {
		t.Fatalf("expected active Azure CLI subscription ID, got %q", actual)
	}
}

func TestSubscriptionIDRejectsAmbiguousSubscriptions(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "")
	options := testClientOptions(responseTransport{body: `{
		"value": [
			{"subscriptionId": "first-subscription", "displayName": "First"},
			{"subscriptionId": "second-subscription", "displayName": "Second"}
		]
	}`})

	_, err := subscriptionId(t.Context(), testDependencies(options.Transport, activeAzureCLISubscription("different-subscription")), types.StringNull(), types.StringNull())

	if err == nil || !strings.Contains(err.Error(), "multiple Azure subscriptions found") {
		t.Fatalf("expected ambiguous subscription error, got %v", err)
	}
}

func TestCreateLocationMapsReturnsLocationPagingDiagnostic(t *testing.T) {
	expectedErr := errors.New("location request failed")
	options := testClientOptions(failingTransport{err: expectedErr})
	var diagnostics diag.Diagnostics

	dependencies := testDependencies(options.Transport, activeAzureCLISubscription("different-subscription"))
	subscriptionID, locations := createLocationMapsWithDependencies(t.Context(), dependencies, types.StringValue("00000000-0000-0000-0000-000000000000"), types.StringNull(), &diagnostics)

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
