package acctest

import (
	"context"
	"terraform-provider-namep/internal/provider"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func RequireAzureAuthentication(t *testing.T) {
	t.Helper()

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Skipf("skipping live Azure test: no Azure credential is available: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	_, err = credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		t.Skipf("skipping live Azure test: Azure authentication is unavailable; run 'az login' locally or configure CI credentials: %v", err)
	}
}
