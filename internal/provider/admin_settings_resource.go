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
	ReminderDays         types.List   `tfsdk:"reminder_days"`
	NotifyHour           types.Int64  `tfsdk:"notify_hour"`
	SlackWebhook         types.String `tfsdk:"slack_webhook"`
	AlertOnOverdue       types.Bool   `tfsdk:"alert_on_overdue"`
	AlertOnDelete        types.Bool   `tfsdk:"alert_on_delete"`
	AlertOnReviewOverdue types.Bool   `tfsdk:"alert_on_review_overdue"`
	ReviewCadenceMonths  types.Int64  `tfsdk:"review_cadence_months"`
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
			"alert_on_review_overdue": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to send alerts when resource reviews are overdue.",
			},
			"review_cadence_months": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "How often resources should be reviewed, in months. Valid values: 6, 12, or 24. Set to null to disable periodic reviews.",
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

	var reviewCadenceMonths *int
	if !m.ReviewCadenceMonths.IsNull() && !m.ReviewCadenceMonths.IsUnknown() {
		v := int(m.ReviewCadenceMonths.ValueInt64())
		reviewCadenceMonths = &v
	}

	return &AdminSettingsRequest{
		ReminderDays:         reminderDays,
		NotifyHour:           int(m.NotifyHour.ValueInt64()),
		SlackWebhook:         m.SlackWebhook.ValueStringPointer(),
		AlertOnOverdue:       m.AlertOnOverdue.ValueBool(),
		AlertOnDelete:        m.AlertOnDelete.ValueBool(),
		AlertOnReviewOverdue: m.AlertOnReviewOverdue.ValueBool(),
		ReviewCadenceMonths:  reviewCadenceMonths,
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

	var slackWebhook types.String
	if apiResp.SlackWebhook != nil {
		slackWebhook = types.StringValue(*apiResp.SlackWebhook)
	} else {
		slackWebhook = types.StringNull()
	}

	var reviewCadenceMonths types.Int64
	if apiResp.ReviewCadenceMonths != nil {
		reviewCadenceMonths = types.Int64Value(int64(*apiResp.ReviewCadenceMonths))
	} else {
		reviewCadenceMonths = types.Int64Null()
	}

	return &AdminSettingsModel{
		ReminderDays:         reminderDaysList,
		NotifyHour:           types.Int64Value(int64(apiResp.NotifyHour)),
		SlackWebhook:         slackWebhook,
		AlertOnOverdue:       types.BoolValue(apiResp.AlertOnOverdue),
		AlertOnDelete:        types.BoolValue(apiResp.AlertOnDelete),
		AlertOnReviewOverdue: types.BoolValue(apiResp.AlertOnReviewOverdue),
		ReviewCadenceMonths:  reviewCadenceMonths,
	}, nil
}
