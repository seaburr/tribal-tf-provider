package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &TribalResourceResource{}
var _ resource.ResourceWithImportState = &TribalResourceResource{}

type TribalResourceResource struct {
	client *TribalClient
}

type TribalResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	DRI                    types.String `tfsdk:"dri"`
	Type                   types.String `tfsdk:"type"`
	ExpirationDate         types.String `tfsdk:"expiration_date"`
	DoesNotExpire          types.Bool   `tfsdk:"does_not_expire"`
	Purpose                types.String `tfsdk:"purpose"`
	GenerationInstructions types.String `tfsdk:"generation_instructions"`
	SecretManagerLink      types.String `tfsdk:"secret_manager_link"`
	SlackWebhook           types.String `tfsdk:"slack_webhook"`
	TeamID                 types.Int64  `tfsdk:"team_id"`
	PublicKeyPEM           types.String `tfsdk:"public_key_pem"`
	CertificateURL         types.String `tfsdk:"certificate_url"`
	AutoRefreshExpiry      types.Bool   `tfsdk:"auto_refresh_expiry"`
	LastReviewedAt         types.String `tfsdk:"last_reviewed_at"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
}

func NewTribalResourceResource() resource.Resource {
	return &TribalResourceResource{}
}

func (r *TribalResourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *TribalResourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Tribal tracked resource (certificate, API key, SSH key, etc.).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric identifier of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the resource.",
			},
			"dri": schema.StringAttribute{
				Required:    true,
				Description: "Directly Responsible Individual for this resource.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of resource: Certificate, API Key, SSH Key, or Other.",
			},
			"expiration_date": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Expiration date in YYYY-MM-DD format. Required unless does_not_expire is true.",
			},
			"does_not_expire": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, the resource does not expire and expiration_date is cleared.",
			},
			"purpose": schema.StringAttribute{
				Required:    true,
				Description: "Purpose/description of the resource.",
			},
			"generation_instructions": schema.StringAttribute{
				Required:    true,
				Description: "Instructions for generating/renewing this resource.",
			},
			"secret_manager_link": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL link to the secret in a secret manager.",
			},
			"slack_webhook": schema.StringAttribute{
				Required:    true,
				Description: "Slack webhook URL for expiration notifications.",
			},
			"team_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the team this resource belongs to. Auto-assigned to the default team if omitted.",
			},
			"public_key_pem": schema.StringAttribute{
				Computed:    true,
				Description: "PEM-encoded public certificate (if uploaded).",
			},
			"certificate_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL of the TLS endpoint to poll for automatic certificate expiry refresh.",
			},
			"auto_refresh_expiry": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, the expiration_date is automatically updated by polling the certificate_url.",
			},
			"last_reviewed_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the resource was last reviewed.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the resource was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the resource was last updated.",
			},
		},
	}
}

func (r *TribalResourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*TribalClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *TribalClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func resourceModelFromResponse(apiResp *ResourceResponse) TribalResourceModel {
	m := TribalResourceModel{
		ID:                     types.StringValue(strconv.Itoa(apiResp.ID)),
		Name:                   types.StringValue(apiResp.Name),
		DRI:                    types.StringValue(apiResp.DRI),
		Type:                   types.StringValue(apiResp.Type),
		DoesNotExpire:          types.BoolValue(apiResp.DoesNotExpire),
		Purpose:                types.StringValue(apiResp.Purpose),
		GenerationInstructions: types.StringValue(apiResp.GenerationInstructions),
		SlackWebhook:           types.StringValue(apiResp.SlackWebhook),
		AutoRefreshExpiry:      types.BoolValue(apiResp.AutoRefreshExpiry),
		CreatedAt:              types.StringValue(apiResp.CreatedAt),
		UpdatedAt:              types.StringValue(apiResp.UpdatedAt),
	}
	if apiResp.ExpirationDate != nil {
		m.ExpirationDate = types.StringValue(*apiResp.ExpirationDate)
	} else {
		m.ExpirationDate = types.StringNull()
	}
	if apiResp.SecretManagerLink != nil {
		m.SecretManagerLink = types.StringValue(*apiResp.SecretManagerLink)
	} else {
		m.SecretManagerLink = types.StringNull()
	}
	if apiResp.TeamID != nil {
		m.TeamID = types.Int64Value(int64(*apiResp.TeamID))
	} else {
		m.TeamID = types.Int64Null()
	}
	if apiResp.PublicKeyPEM != nil {
		m.PublicKeyPEM = types.StringValue(*apiResp.PublicKeyPEM)
	} else {
		m.PublicKeyPEM = types.StringNull()
	}
	if apiResp.CertificateURL != nil {
		m.CertificateURL = types.StringValue(*apiResp.CertificateURL)
	} else {
		m.CertificateURL = types.StringNull()
	}
	if apiResp.LastReviewedAt != nil {
		m.LastReviewedAt = types.StringValue(*apiResp.LastReviewedAt)
	} else {
		m.LastReviewedAt = types.StringNull()
	}
	return m
}

func planToResourceRequest(plan TribalResourceModel) ResourceRequest {
	req := ResourceRequest{
		Name:                   plan.Name.ValueString(),
		DRI:                    plan.DRI.ValueString(),
		Type:                   plan.Type.ValueString(),
		DoesNotExpire:          plan.DoesNotExpire.ValueBool(),
		Purpose:                plan.Purpose.ValueString(),
		GenerationInstructions: plan.GenerationInstructions.ValueString(),
		SlackWebhook:           plan.SlackWebhook.ValueString(),
		AutoRefreshExpiry:      plan.AutoRefreshExpiry.ValueBool(),
	}
	if !plan.ExpirationDate.IsNull() && !plan.ExpirationDate.IsUnknown() {
		v := plan.ExpirationDate.ValueString()
		req.ExpirationDate = &v
	}
	if !plan.SecretManagerLink.IsNull() && !plan.SecretManagerLink.IsUnknown() {
		req.SecretManagerLink = plan.SecretManagerLink.ValueString()
	}
	if !plan.TeamID.IsNull() && !plan.TeamID.IsUnknown() {
		v := int(plan.TeamID.ValueInt64())
		req.TeamID = &v
	}
	if !plan.CertificateURL.IsNull() && !plan.CertificateURL.IsUnknown() {
		req.CertificateURL = plan.CertificateURL.ValueString()
	}
	return req
}

func (r *TribalResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TribalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.CreateResource(planToResourceRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Resource", err.Error())
		return
	}

	diags = resp.State.Set(ctx, resourceModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

func (r *TribalResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TribalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	apiResp, err := r.client.GetResource(id)
	if err != nil {
		if strings.Contains(err.Error(), "API error 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Resource", err.Error())
		return
	}

	diags = resp.State.Set(ctx, resourceModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

func (r *TribalResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TribalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state TribalResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	apiResp, err := r.client.UpdateResource(id, planToResourceRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Resource", err.Error())
		return
	}

	diags = resp.State.Set(ctx, resourceModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

func (r *TribalResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TribalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	if err := r.client.DeleteResource(id); err != nil {
		if strings.Contains(err.Error(), "API error 404") {
			return
		}
		resp.Diagnostics.AddError("Error Deleting Resource", err.Error())
	}
}

func (r *TribalResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if _, err := strconv.Atoi(id); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be a numeric resource ID")
		return
	}

	apiResp, err := r.client.GetResource(mustAtoi(id))
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Resource", err.Error())
		return
	}

	diags := resp.State.Set(ctx, resourceModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
