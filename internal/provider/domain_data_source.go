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
	_ datasource.DataSource              = &DomainDataSource{}
	_ datasource.DataSourceWithConfigure = &DomainDataSource{}
)

// NewDomainDataSource is a helper function to simplify the provider implementation.
func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

// DomainDataSource is the data source implementation.
type DomainDataSource struct {
	client *loopia.API
}

// DomainDataSourceModel maps the data source schema data.
type DomainDataSourceModel struct {
	Name            types.String `tfsdk:"name"`
	Paid            types.Bool   `tfsdk:"paid"`
	Registered      types.Bool   `tfsdk:"registered"`
	RenewalStatus   types.String `tfsdk:"renewal_status"`
	ExpirationDate  types.String `tfsdk:"expiration_date"`
	ReferenceNumber types.Int32  `tfsdk:"reference_number"`
}

// Metadata returns the data source type name.
func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the data source.
func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific domain.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The domain name to retrieve data for",
			},
			"paid": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the domain is paid for",
			},
			"registered": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the domain is registered",
			},
			"renewal_status": schema.StringAttribute{
				Computed:    true,
				Description: "The renewal status of the domain",
			},
			"expiration_date": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration date of the domain",
			},
			"reference_number": schema.Int32Attribute{
				Computed:    true,
				Description: "The reference number of the domain",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DomainDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get domain details from API
	domain, err := d.client.GetDomain(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Loopia Domain",
			err.Error(),
		)
		return
	}

	// Map response directly to state
	state.Name = types.StringValue(domain.Name)
	state.Paid = types.BoolValue(domain.Paid)
	state.Registered = types.BoolValue(domain.Registered)
	state.RenewalStatus = types.StringValue(domain.RenewalStatus)
	state.ExpirationDate = types.StringValue(domain.ExpirationDate)
	state.ReferenceNumber = types.Int32Value(domain.ReferenceNumber)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Configure adds the provider configured client to the data source.
func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
