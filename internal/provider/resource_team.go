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

var _ resource.Resource = &TribalTeamResource{}
var _ resource.ResourceWithImportState = &TribalTeamResource{}

type TribalTeamResource struct {
	client *TribalClient
}

type TribalTeamModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewTribalTeamResource() resource.Resource {
	return &TribalTeamResource{}
}

func (r *TribalTeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *TribalTeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Tribal singleton team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric identifier of the team.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the team.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the team was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TribalTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func teamModelFromResponse(apiResp *TeamResponse) TribalTeamModel {
	return TribalTeamModel{
		ID:        types.StringValue(strconv.Itoa(apiResp.ID)),
		Name:      types.StringValue(apiResp.Name),
		CreatedAt: types.StringValue(apiResp.CreatedAt),
	}
}

func (r *TribalTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TribalTeamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.CreateTeam(TeamRequest{Name: plan.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Team", err.Error())
		return
	}

	diags = resp.State.Set(ctx, teamModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

func (r *TribalTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TribalTeamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Team ID", err.Error())
		return
	}

	apiResp, err := r.client.GetTeam(id)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Team", err.Error())
		return
	}

	diags = resp.State.Set(ctx, teamModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

func (r *TribalTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TribalTeamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state TribalTeamModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Team ID", err.Error())
		return
	}

	apiResp, err := r.client.UpdateTeam(id, TeamRequest{Name: plan.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Team", err.Error())
		return
	}

	diags = resp.State.Set(ctx, teamModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}

// Delete is a no-op — the API has no DELETE /admin/teams/{id} endpoint.
func (r *TribalTeamResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *TribalTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be a numeric team ID")
		return
	}

	apiResp, err := r.client.GetTeam(id)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Team", err.Error())
		return
	}

	diags := resp.State.Set(ctx, teamModelFromResponse(apiResp))
	resp.Diagnostics.Append(diags...)
}
