// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/diskoteket/loopia-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure LoopiaProvider satisfies various provider interfaces.
var _ provider.Provider = &LoopiaProvider{}

//var _ provider.ProviderWithFunctions = &LoopiaProvider{}
//var _ provider.ProviderWithEphemeralResources = &LoopiaProvider{}

// LoopiaProvider defines the provider implementation.
type LoopiaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// LoopiaProviderModel describes the provider data model.
type loopiaProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *LoopiaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "loopia"
	resp.Version = p.version
}

func (p *LoopiaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "The user name to use for Loopia API authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The user password to use for Loopia API authentication",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *LoopiaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Loopia Client")

	// Retrieve provider data from configuration
	var config loopiaProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Loopia API Username",
			"The provider cannot create the Loopia API client as there is an unknown configuration value for the Loopia API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LOOPIA_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Loopia API Password",
			"The provider cannot create the Loopia API client as there is an unknown configuration value for the Loopia API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LOOPIA_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	username := os.Getenv("LOOPIA_USERNAME")
	password := os.Getenv("LOOPIA_PASSWORD")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Loopia API Username",
			"The provider cannot create the Loopia API client as there is a missing or empty value for the Loopia API username. "+
				"Set the username value in the configuration or use the LOOPIA_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Loopia API Password",
			"The provider cannot create the Loopia API client as there is a missing or empty value for the Loopia API password. "+
				"Set the password value in the configuration or use the LOOPIA_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "loopia_username", username)
	ctx = tflog.SetField(ctx, "loopia_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "loopia_password")

	tflog.Debug(ctx, "Creating Loopia Client")

	// Create a new Loopia client using the configuration values
	client, err := loopia.New(username, password)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Loopia API Client",
			"An unexpected error occurred when creating the Loopia API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Loopia Client Error: "+err.Error(),
		)
		return
	}

	// Make the Loopia client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Loopia client", map[string]any{"success": true})
}

func (p *LoopiaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubdomainResource,
		NewZoneRecordResource,
	}
}

//func (p *LoopiaProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
//	return []func() ephemeral.EphemeralResource{
//		NewExampleEphemeralResource,
//	}
//}

func (p *LoopiaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
		NewDomainsDataSource,
		NewSubdomainsDataSource,
		NewZoneRecordsDataSource,
	}
}

//func (p *LoopiaProvider) Functions(ctx context.Context) []func() function.Function {
//	return []func() function.Function{
//		NewExampleFunction,
//	}
//}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LoopiaProvider{
			version: version,
		}
	}
}
