package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AdminSettingsResource{}

type AdminSettingsResource struct {
	client *TribalClient
}

type AdminSettingsModel struct {
	OrgName        types.String `tfsdk:"org_name"`
	ReminderDays   types.List   `tfsdk:"reminder_days"`
	NotifyHour     types.Int64  `tfsdk:"notify_hour"`
	SlackWebhook   types.String `tfsdk:"slack_webhook"`
	AlertOnOverdue types.Bool   `tfsdk:"alert_on_overdue"`
	AlertOnDelete  types.Bool   `tfsdk:"alert_on_delete"`
}

func NewAdminSettingsResource() resource.Resource {
	return &AdminSettingsResource{}
}

func (r *AdminSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_settings"
}

func (r *AdminSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Tribal organization-wide admin settings. This is a singleton resource.",
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Organization name displayed in the Tribal UI.",
			},
			"reminder_days": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "List of days before expiration to send reminders (e.g., [30, 14, 7]).",
			},
			"notify_hour": schema.Int64Attribute{
				Required:    true,
				Description: "UTC hour (0-23) at which to send daily reminders.",
			},
			"slack_webhook": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Organization-wide Slack webhook URL for expiration notifications.",
			},
			"alert_on_overdue": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to send alerts for overdue (already expired) resources.",
			},
			"alert_on_delete": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to send an admin Slack alert when a resource is deleted.",
			},
		},
	}
}

func (r *AdminSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AdminSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AdminSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, err := r.modelToRequest(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Building Request", err.Error())
		return
	}

	apiResp, err := r.client.UpdateAdminSettings(*updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Configuring Admin Settings", err.Error())
		return
	}

	newState, diagErr := r.responseToModel(ctx, apiResp)
	resp.Diagnostics.Append(diagErr...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *AdminSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	apiResp, err := r.client.GetAdminSettings()
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Admin Settings", err.Error())
		return
	}

	newState, diagErr := r.responseToModel(ctx, apiResp)
	resp.Diagnostics.Append(diagErr...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *AdminSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AdminSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, err := r.modelToRequest(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Building Request", err.Error())
		return
	}

	apiResp, err := r.client.UpdateAdminSettings(*updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Admin Settings", err.Error())
		return
	}

	newState, diagErr := r.responseToModel(ctx, apiResp)
	resp.Diagnostics.Append(diagErr...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

// Delete is a no-op for admin settings since it's a singleton.
func (r *AdminSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *AdminSettingsResource) modelToRequest(ctx context.Context, m AdminSettingsModel) (*AdminSettingsRequest, error) {
	var days []int64
	diags := m.ReminderDays.ElementsAs(ctx, &days, false)
	if diags.HasError() {
		return nil, fmt.Errorf("reading reminder_days: %s", diags[0].Detail())
	}

	reminderDays := make([]int, len(days))
	for i, d := range days {
		reminderDays[i] = int(d)
	}

	return &AdminSettingsRequest{
		OrgName:        m.OrgName.ValueString(),
		ReminderDays:   reminderDays,
		NotifyHour:     int(m.NotifyHour.ValueInt64()),
		SlackWebhook:   m.SlackWebhook.ValueString(),
		AlertOnOverdue: m.AlertOnOverdue.ValueBool(),
		AlertOnDelete:  m.AlertOnDelete.ValueBool(),
	}, nil
}

func (r *AdminSettingsResource) responseToModel(ctx context.Context, apiResp *AdminSettingsResponse) (*AdminSettingsModel, diag.Diagnostics) {
	reminderDayValues := make([]attr.Value, len(apiResp.ReminderDays))
	for i, d := range apiResp.ReminderDays {
		reminderDayValues[i] = types.Int64Value(int64(d))
	}
	reminderDaysList, diags := types.ListValue(types.Int64Type, reminderDayValues)
	if diags.HasError() {
		return nil, diags
	}

	var orgName types.String
	if apiResp.OrgName != nil {
		orgName = types.StringValue(*apiResp.OrgName)
	} else {
		orgName = types.StringNull()
	}

	var slackWebhook types.String
	if apiResp.SlackWebhook != nil {
		slackWebhook = types.StringValue(*apiResp.SlackWebhook)
	} else {
		slackWebhook = types.StringNull()
	}

	return &AdminSettingsModel{
		OrgName:        orgName,
		ReminderDays:   reminderDaysList,
		NotifyHour:     types.Int64Value(int64(apiResp.NotifyHour)),
		SlackWebhook:   slackWebhook,
		AlertOnOverdue: types.BoolValue(apiResp.AlertOnOverdue),
		AlertOnDelete:  types.BoolValue(apiResp.AlertOnDelete),
	}, nil
}
