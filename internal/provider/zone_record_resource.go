// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/diskoteket/loopia-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &zoneRecordResource{}
	_ resource.ResourceWithConfigure = &zoneRecordResource{}
)

// NewZoneRecordResource is a helper function to simplify the provider implementation.
func NewZoneRecordResource() resource.Resource {
	return &zoneRecordResource{}
}

// zoneRecordResource is the resource implementation.
type zoneRecordResource struct {
	client *loopia.API
}

// ZoneRecordResourceModel maps the resource schema data.
type ZoneRecordResourceModel struct {
	Domain    types.String `tfsdk:"domain"`
	Subdomain types.String `tfsdk:"subdomain"`
	Record    recordModel  `tfsdk:"record"`
}

type recordModel struct {
	Type     types.String `tfsdk:"type"`
	Ttl      types.Int32  `tfsdk:"ttl"`
	Priority types.Int32  `tfsdk:"priority"`
	Value    types.String `tfsdk:"value"`
	RecordId types.Int32  `tfsdk:"record_id"`
}

// toClientRecord converts the Terraform model to a Loopia API record.
func (m *recordModel) toClientRecord() loopia.Record {
	return loopia.Record{
		ID:       int64(m.RecordId.ValueInt32()),
		TTL:      int(m.Ttl.ValueInt32()),
		Type:     m.Type.ValueString(),
		Value:    m.Value.ValueString(),
		Priority: int(m.Priority.ValueInt32()),
	}
}

// recordModelFromClient converts a Loopia API record to the Terraform model.
func recordModelFromClient(rec loopia.Record) recordModel {
	return recordModel{
		Type:     types.StringValue(rec.Type),
		Ttl:      types.Int32Value(int32(rec.TTL)),
		Priority: types.Int32Value(int32(rec.Priority)),
		Value:    types.StringValue(rec.Value),
		RecordId: types.Int32Value(int32(rec.ID)),
	}
}

// recordsMatch checks if a client record matches the planned record values.
func (r *zoneRecordResource) recordsMatch(apiRecord loopia.Record, planRecord recordModel) bool {
	return apiRecord.Type == planRecord.Type.ValueString() &&
		apiRecord.Value == planRecord.Value.ValueString() &&
		apiRecord.TTL == int(planRecord.Ttl.ValueInt32()) &&
		apiRecord.Priority == int(planRecord.Priority.ValueInt32())
}

// Metadata returns the resource type name.
func (r *zoneRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone_record"
}

// Schema defines the schema for the resource.
func (r *zoneRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS zone record in Loopia.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain name to create records for.",
			},
			"subdomain": schema.StringAttribute{
				Required:    true,
				Description: "The subdomain to create records for.",
			},
			"record": schema.SingleNestedAttribute{
				Description: "The DNS record to manage.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "The type of the record (e.g., 'A', 'CNAME', 'MX').",
						Required:    true,
					},
					"value": schema.StringAttribute{
						Description: "The value of the record. For an 'A' record, this is an IPv4 address.",
						Required:    true,
					},
					"ttl": schema.Int32Attribute{
						Description: "Time-to-live for the record in seconds.",
						Optional:    true,
						Computed:    true,
					},
					"priority": schema.Int32Attribute{
						Description: "The priority for MX records.",
						Optional:    true,
						Computed:    true,
					},
					"record_id": schema.Int32Attribute{
						Description: "The unique identifier for the record (computed).",
						Computed:    true,
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *zoneRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ZoneRecordResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the record using the API
	planRecord := plan.Record.toClientRecord()
	err := r.client.AddZoneRecord(
		plan.Domain.ValueString(),
		plan.Subdomain.ValueString(),
		&planRecord,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Zone Record",
			fmt.Sprintf("Could not create zone record: %s", err.Error()),
		)
		return
	}

	// Fetch all records to find the newly created one with its ID
	records, err := r.client.GetZoneRecords(
		plan.Domain.ValueString(),
		plan.Subdomain.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Fetching Zone Records After Creation",
			fmt.Sprintf("Could not list zone records: %s", err.Error()),
		)
		return
	}

	// Find the matching record
	var createdRecord *loopia.Record
	for _, rec := range records {
		if r.recordsMatch(rec, plan.Record) {
			createdRecord = &rec
			break
		}
	}

	if createdRecord == nil {
		resp.Diagnostics.AddError(
			"Unable to Identify Created Record",
			"Could not find the newly created record in the API response.",
		)
		return
	}

	// Update plan with the record including its ID
	plan.Record = recordModelFromClient(*createdRecord)

	// Save state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *zoneRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ZoneRecordResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the record from the API
	rec, err := r.client.GetZoneRecord(
		state.Domain.ValueString(),
		state.Subdomain.ValueString(),
		int64(state.Record.RecordId.ValueInt32()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zone Record",
			fmt.Sprintf("Could not read zone record ID %d: %s",
				state.Record.RecordId.ValueInt32(), err.Error()),
		)
		return
	}

	// Update state with fresh data
	state.Record = recordModelFromClient(*rec)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *zoneRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ZoneRecordResourceModel

	// Get current state to retrieve the record ID
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get planned values
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the record ID from state for the update
	plan.Record.RecordId = state.Record.RecordId

	// Update the record via API
	rec := plan.Record.toClientRecord()
	_, err := r.client.UpdateZoneRecord(
		plan.Domain.ValueString(),
		plan.Subdomain.ValueString(),
		rec,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Zone Record",
			fmt.Sprintf("Could not update zone record ID %d: %s",
				plan.Record.RecordId.ValueInt32(), err.Error()),
		)
		return
	}

	// Refresh the state with updated data
	updatedRec, err := r.client.GetZoneRecord(
		plan.Domain.ValueString(),
		plan.Subdomain.ValueString(),
		int64(plan.Record.RecordId.ValueInt32()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Refreshing Zone Record After Update",
			fmt.Sprintf("Could not read updated zone record: %s", err.Error()),
		)
		return
	}

	plan.Record = recordModelFromClient(*updatedRec)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *zoneRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ZoneRecordResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the record via API
	_, err := r.client.RemoveZoneRecord(
		state.Domain.ValueString(),
		state.Subdomain.ValueString(),
		int64(state.Record.RecordId.ValueInt32()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Zone Record",
			fmt.Sprintf("Could not delete zone record ID %d: %s",
				state.Record.RecordId.ValueInt32(), err.Error()),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *zoneRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*loopia.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *loopia.API, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}
