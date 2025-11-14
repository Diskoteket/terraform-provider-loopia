// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/diskoteket/loopia-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &zoneRecordsDataSource{}
	_ datasource.DataSourceWithConfigure = &zoneRecordsDataSource{}
)

// NewZoneRecordsDataSource is a helper function to simplify the provider implementation.
func NewZoneRecordsDataSource() datasource.DataSource {
	return &zoneRecordsDataSource{}
}

// zoneRecordsDataSource is the data source implementation.
type zoneRecordsDataSource struct {
	client *loopia.API
}

// ZoneRecordsDataSourceModel maps the data source schema data.
type ZoneRecordsDataSourceModel struct {
	Domain      types.String       `tfsdk:"domain"`
	Subdomain   types.String       `tfsdk:"subdomain"`
	ZoneRecords []zoneRecordsModel `tfsdk:"zone_records"`
}

type zoneRecordsModel struct {
	Type     types.String `tfsdk:"type"`
	Ttl      types.Int32  `tfsdk:"ttl"`
	Priority types.Int32  `tfsdk:"priority"`
	Value    types.String `tfsdk:"value"`
	RecordId types.Int32  `tfsdk:"record_id"`
}

// Metadata returns the data source type name.
func (d *zoneRecordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone_records"
}

// Schema defines the schema for the data source.
func (d *zoneRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches DNS zone records for a specific domain and subdomain",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to retrieve records for.",
				Required:    true,
			},
			"subdomain": schema.StringAttribute{
				Description: "The subdomain to retrieve records for.",
				Required:    true,
			},
			"zone_records": schema.ListNestedAttribute{
				Description: "List of DNS zone records.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of the record (e.g., 'A', 'CNAME', 'MX').",
							Computed:    true,
						},
						"ttl": schema.Int32Attribute{
							Description: "Time-to-live for the record in seconds.",
							Computed:    true,
						},
						"priority": schema.Int32Attribute{
							Description: "The priority for MX records.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The value of the record. For an 'A' record, this is an IPv4 address.",
							Computed:    true,
						},
						"record_id": schema.Int32Attribute{
							Description: "The unique identifier for the record.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *zoneRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ZoneRecordsDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneRecords, err := d.client.GetZoneRecords(
		state.Domain.ValueString(),
		state.Subdomain.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Loopia Zone Records",
			fmt.Sprintf("Could not read zone records: %s", err.Error()),
		)
		return
	}

	// Map response to model
	for _, record := range zoneRecords {
		zoneRecordState := zoneRecordsModel{
			Type:     types.StringValue(record.Type),
			Ttl:      types.Int32Value(int32(record.TTL)),
			Priority: types.Int32Value(int32(record.Priority)),
			Value:    types.StringValue(record.Value),
			RecordId: types.Int32Value(int32(record.ID)),
		}
		state.ZoneRecords = append(state.ZoneRecords, zoneRecordState)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *zoneRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*loopia.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *loopia.API, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}
