package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &TribalProvider{}

type TribalProvider struct {
	version string
}

type TribalProviderModel struct {
	Host   types.String `tfsdk:"host"`
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TribalProvider{
			version: version,
		}
	}
}

func (p *TribalProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tribal"
	resp.Version = p.version
}

func (p *TribalProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing Tribal application resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "The base URL of the Tribal API (e.g., http://localhost:8000). Defaults to TRIBAL_HOST environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API key for authenticating with the Tribal API. Defaults to TRIBAL_API_KEY environment variable.",
			},
		},
	}
}

func (p *TribalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TribalProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("TRIBAL_HOST")
	if !config.Host.IsNull() && !config.Host.IsUnknown() {
		host = config.Host.ValueString()
	}
	if host == "" {
		host = "http://localhost:8000"
	}

	apiKey := os.Getenv("TRIBAL_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The Tribal provider requires an API key. Set it via the api_key attribute or TRIBAL_API_KEY environment variable.",
		)
		return
	}

	client := NewTribalClient(host, apiKey)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TribalProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTribalResourceResource,
		NewAdminSettingsResource,
		NewTribalTeamResource,
	}
}

func (p *TribalProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
