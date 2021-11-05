package generate_test

import (
	"testing"

	"github.com/lacework/go-sdk/generate"
	"github.com/stretchr/testify/assert"
)

func TestGenerationCloudTrail(t *testing.T) {
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportCt, hcl)
}

func TestGenerationConfig(t *testing.T) {
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureConfig: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportConfig, hcl)
}

func TestGenerationConfigAndCt(t *testing.T) {
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureConfig:     true,
		ConfigureCloudtrail: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportConfig+moduleImportCt, hcl)
}

func TestGenerationWithProviderRegion(t *testing.T) {
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		CloudTrailRegion: "us-east-2",
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+awsProvider, hcl)
}

var requiredProviders = `terraform {
  required_providers {
    lacework = {
      source  = "lacework/lacework"
      version = "~> 0.3"
    }
  }
}

`

var awsProvider = `provider "aws" {
  region = "us-east-2"
}

`

var moduleImportCt = `module "main_cloudtrail" {
  iam_role_arn          = module.aws_config.iam_role_arn
  iam_role_external_id  = module.aws_config.external_id
  iam_role_name         = module.aws_config.iam_role_name
  source                = "lacework/cloudtrail/aws"
  use_existing_iam_role = true
  version               = "~> 0.1"
}

`

var moduleImportConfig = `module "aws_config" {
  source  = "lacework/config/aws"
  version = "~> 0.1"
}

`
