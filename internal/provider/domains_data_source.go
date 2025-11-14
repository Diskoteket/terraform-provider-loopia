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
	_ datasource.DataSource              = &DomainsDataSource{}
	_ datasource.DataSourceWithConfigure = &DomainsDataSource{}
)

// NewDomainsDataSource is a helper function to simplify the provider implementation.
func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

// DomainsDataSource is the data source implementation.
type DomainsDataSource struct {
	client *loopia.API
}

// DomainsDataSourceModel maps the data source schema data.
type DomainsDataSourceModel struct {
	Domains []domainsModel `tfsdk:"domains"`
}

// domainsModel maps coffees schema data.
type domainsModel struct {
	Name            types.String `tfsdk:"name"`
	Paid            types.Bool   `tfsdk:"paid"`
	Registered      types.Bool   `tfsdk:"registered"`
	RenewalStatus   types.String `tfsdk:"renewal_status"`
	ExpirationDate  types.String `tfsdk:"expiration_date"`
	ReferenceNumber types.Int32  `tfsdk:"reference_number"`
}

// Metadata returns the data source type name.
func (d *DomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

// Schema defines the schema for the data source.
func (d *DomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"paid": schema.BoolAttribute{
							Computed: true,
						},
						"registered": schema.BoolAttribute{
							Computed: true,
						},
						"renewal_status": schema.StringAttribute{
							Computed: true,
						},
						"expiration_date": schema.StringAttribute{
							Computed: true,
						},
						"reference_number": schema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DomainsDataSourceModel

	domains, err := d.client.GetDomains()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Loopia Domains",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, domain := range domains {
		domainState := domainsModel{
			Name:            types.StringValue(domain.Name),
			Paid:            types.BoolValue(domain.Paid),
			Registered:      types.BoolValue(domain.Registered),
			RenewalStatus:   types.StringValue(domain.RenewalStatus),
			ExpirationDate:  types.StringValue(domain.ExpirationDate),
			ReferenceNumber: types.Int32Value(domain.ReferenceNumber),
		}
		state.Domains = append(state.Domains, domainState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *DomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
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
