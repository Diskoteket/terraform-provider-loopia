// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/diskoteket/loopia-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &SubdomainsDataSource{}
	_ datasource.DataSourceWithConfigure = &SubdomainsDataSource{}
)

// NewSubdomainsDataSource is a helper function to simplify the provider implementation.
func NewSubdomainsDataSource() datasource.DataSource {
	return &SubdomainsDataSource{}
}

// SubdomainsDataSource is the data source implementation.
type SubdomainsDataSource struct {
	client *loopia.API
}

// SubdomainsDataSourceModel maps the data source schema data.
type SubdomainsDataSourceModel struct {
	Domain     types.String `tfsdk:"domain"`
	Subdomains types.List   `tfsdk:"subdomains"`
}

// Metadata returns the data source type name.
func (d *SubdomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomains"
}

// Schema defines the schema for the data source.
func (d *SubdomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about all subdomains.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to retrieve subdomains for",
				Required:    true,
			},
			"subdomains": schema.ListAttribute{
				Description: "List of subdomain names",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *SubdomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SubdomainsDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subdomains, err := d.client.GetSubdomains(state.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Loopia Subdomains",
			err.Error(),
		)
		return
	}

	// Convert subdomain names to a list of strings
	subdomainNames := make([]attr.Value, 0, len(subdomains))
	for _, subdomain := range subdomains {
		subdomainNames = append(subdomainNames, types.StringValue(subdomain.Name))
	}

	// Create a List from the string values
	listValue, diags := types.ListValue(types.StringType, subdomainNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Subdomains = listValue

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *SubdomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
