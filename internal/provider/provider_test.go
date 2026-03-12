package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	provider "github.com/cberndt/tribal-tf-provider/internal/provider"
)

const (
	testAPIKey = "tribal_sk_cbf8c9f5b9a290dad3b51434eb0738eff8d9abcf382d765194b46c96ad13c225"
	testHost   = "http://localhost:8000"
)

func testProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"tribal": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func providerConfig() string {
	return `
provider "tribal" {
  host    = "` + testHost + `"
  api_key = "` + testAPIKey + `"
}
`
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestAccProvider verifies the provider initializes and can manage admin settings.
func TestAccProvider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "tribal_admin_settings" "smoke" {
  reminder_days    = [30, 7]
  notify_hour      = 9
  alert_on_overdue = false
  alert_on_delete  = false
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tribal_admin_settings.smoke", "notify_hour", "9"),
				),
			},
		},
	})
}
