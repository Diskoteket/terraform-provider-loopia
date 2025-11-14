// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/diskoteket/loopia-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &subdomainResource{}
	_ resource.ResourceWithConfigure = &subdomainResource{}
)

// NewSubdomainResource is a helper function to simplify the provider implementation.
func NewSubdomainResource() resource.Resource {
	return &subdomainResource{}
}

// subdomainResource is the resource implementation.
type subdomainResource struct {
	client *loopia.API
}

// SubdomainsDataSourceModel maps the data source schema data.
type SubdomainResourceModel struct {
	Domain    types.String `tfsdk:"domain"`
	Subdomain types.String `tfsdk:"subdomain"`
}

// Metadata returns the resource type name.
func (r *subdomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomain"
}

// Schema defines the schema for the resource.
func (r *subdomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to create the subdomain for",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subdomain": schema.StringAttribute{
				Description: "The subdomain to create",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *subdomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from Plan
	var plan SubdomainResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We do not need to generate a request because we got everyhting we need...

	// Get domain details from API
	_, err := r.client.AddSubdomain(plan.Domain.ValueString(), plan.Subdomain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subdomain",
			"Could not create subdomain, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response directly to state
	//state.Domain = types.StringValue()

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
//
// Since the Loopia API lacks a getSubdomain method we use getSubdomains,
// we then iterate through the list to see if the subdomain is there or not.
func (r *subdomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubdomainResourceModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch all subdomains for the domain
	subdomains, err := r.client.GetSubdomains(state.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Subdomains",
			err.Error(),
		)
		return
	}

	// Check if our subdomain exists
	found := false
	for _, subdomain := range subdomains {
		if subdomain.Name == state.Subdomain.ValueString() {
			found = true
			break
		}
	}

	if !found {
		// Subdomain no longer exists, remove resource from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Subdomain exists, keep the state as-is
	resp.State.Set(ctx, &state)
}

// Update updates the resource and sets the updated Terraform state on success.
//
// Since the Loopia API lacks an update method for subdomains we trigger recreation on all changes to the attributes.
// This is just a placeholder.
func (r *subdomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subdomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubdomainResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing subdomain
	_, err := r.client.RemoveSubDomain(state.Domain.ValueString(), state.Subdomain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Loopia Subdomain",
			"Could not delete subdomain, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *subdomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*loopia.API)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
