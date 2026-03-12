package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTribalResource_basic(t *testing.T) {
	resourceName := "tribal_resource.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testTribalResourceConfig("Test TF Resource", "alice@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test TF Resource"),
					resource.TestCheckResourceAttr(resourceName, "dri", "alice@example.com"),
					resource.TestCheckResourceAttr(resourceName, "type", "API Key"),
					resource.TestCheckResourceAttr(resourceName, "expiration_date", "2027-01-01"),
					resource.TestCheckResourceAttr(resourceName, "purpose", "Terraform acceptance test resource"),
					resource.TestCheckResourceAttr(resourceName, "generation_instructions", "Run keygen script"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// ImportState
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: providerConfig() + testTribalResourceConfig("Test TF Resource Updated", "bob@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test TF Resource Updated"),
					resource.TestCheckResourceAttr(resourceName, "dri", "bob@example.com"),
				),
			},
		},
	})
}

func TestAccTribalResource_withSecretManagerLink(t *testing.T) {
	resourceName := "tribal_resource.with_secret"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "tribal_resource" "with_secret" {
  name                    = "TF Resource With Secret"
  dri                     = "devops@example.com"
  type                    = "Certificate"
  expiration_date         = "2027-06-30"
  purpose                 = "Testing secret manager link"
  generation_instructions = "Use certbot to regenerate"
  secret_manager_link     = "https://vault.example.com/secret/my-cert"
  slack_webhook           = "https://hooks.slack.com/services/fake/webhook/url"
}
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "Certificate"),
					resource.TestCheckResourceAttr(resourceName, "secret_manager_link", "https://vault.example.com/secret/my-cert"),
				),
			},
		},
	})
}

func TestAccTribalResource_certAutoRefresh(t *testing.T) {
	resourceName := "tribal_resource.cert"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "tribal_resource" "cert" {
  name                    = "TF Auto-Refresh Cert"
  dri                     = "devops@example.com"
  type                    = "Certificate"
  expiration_date         = "2027-12-31"
  purpose                 = "Testing cert auto-refresh fields"
  generation_instructions = "Use certbot to regenerate"
  slack_webhook           = "https://hooks.slack.com/services/fake/webhook/url"
  certificate_url         = "https://example.com"
  auto_refresh_expiry     = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "certificate_url", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "auto_refresh_expiry", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "team_id"),
				),
			},
		},
	})
}

func testTribalResourceConfig(name, dri string) string {
	return fmt.Sprintf(`
resource "tribal_resource" "test" {
  name                    = %q
  dri                     = %q
  type                    = "API Key"
  expiration_date         = "2027-01-01"
  purpose                 = "Terraform acceptance test resource"
  generation_instructions = "Run keygen script"
  slack_webhook           = "https://hooks.slack.com/services/fake/webhook/url"
}
`, name, dri)
}
