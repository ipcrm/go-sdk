package generate_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lacework/go-sdk/generate"
	"github.com/stretchr/testify/assert"
)

func TestGenerationCloudTrail(t *testing.T) {
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportCtWithoutConfig, hcl)
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
		ConfigureConfig: true, ConfigureCloudtrail: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportConfig+moduleImportCtWithConfig, hcl)
}

func TestGenerationWithProviderRegion(t *testing.T) {
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		AwsRegion: "us-east-2",
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+awsProvider, hcl)
}

func TestGenerationCloudtrailForceDestroyS3(t *testing.T) {
	data := generate.GenerateAwsTfConfiguration{
		Args: &generate.GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:  true,
			ForceDestroyS3Bucket: true,
		},
	}
	data.CreateCloudtrailBlock()
	assert.Equal(t,
		"bucket_force_destroy=true\n",
		string(data.Blocks[0].Body().GetAttribute("bucket_force_destroy").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailConsolidatedTrail(t *testing.T) {
	data := generate.GenerateAwsTfConfiguration{
		Args: &generate.GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:       true,
			UseConsolidatedCloudtrail: true,
		},
	}
	data.CreateCloudtrailBlock()
	assert.Equal(t,
		"consolidated_trail=true\n",
		string(data.Blocks[0].Body().GetAttribute("consolidated_trail").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailExistingSns(t *testing.T) {
	existingSnsTopicArn := "arn:aws:sns:::foo"
	data := generate.GenerateAwsTfConfiguration{
		Args: &generate.GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail: true,
			ExistingSnsTopicArn: existingSnsTopicArn,
		},
	}
	data.CreateCloudtrailBlock()
	assert.Equal(t,
		fmt.Sprintf("sns_topic_arn=\"%s\"\n", existingSnsTopicArn),
		string(data.Blocks[0].Body().GetAttribute("sns_topic_arn").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		"use_existing_sns_topic=true\n",
		string(data.Blocks[0].Body().GetAttribute("use_existing_sns_topic").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailExistingBucket(t *testing.T) {
	existingBucketArn := "arn:aws:s3:::test-bucket-12345"
	data := generate.GenerateAwsTfConfiguration{
		Args: &generate.GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:   true,
			UseExistingCloudtrail: true,
			ExistingBucketArn:     existingBucketArn,
		},
	}
	data.CreateCloudtrailBlock()
	assert.Equal(t,
		"use_existing_cloudtrail=true\n",
		string(data.Blocks[0].Body().GetAttribute("use_existing_cloudtrail").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("bucket_arn=\"%s\"\n", existingBucketArn),
		string(data.Blocks[0].Body().GetAttribute("bucket_arn").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailExistingRole(t *testing.T) {
	iamRoleArn := "arn:aws:iam::123456789012:role/test-role"
	iamRoleName := "test-role"
	extId := "1234567890123456"

	data := generate.GenerateAwsTfConfiguration{
		Args: &generate.GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:       true,
			UseExistingIamRole:        true,
			ExistingIamRoleArn:        iamRoleArn,
			ExistingIamRoleName:       iamRoleName,
			ExistingIamRoleExternalId: extId,
		},
	}
	data.CreateCloudtrailBlock()
	assert.Equal(t,
		"use_existing_iam_role=true\n",
		string(data.Blocks[0].Body().GetAttribute("use_existing_iam_role").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("iam_role_name=\"%s\"\n", iamRoleName),
		string(data.Blocks[0].Body().GetAttribute("iam_role_name").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("iam_role_arn=\"%s\"\n", iamRoleArn),
		string(data.Blocks[0].Body().GetAttribute("iam_role_arn").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("iam_role_external_id=\"%s\"\n", extId),
		string(data.Blocks[0].Body().GetAttribute("iam_role_external_id").BuildTokens(nil).Bytes()))
}

func TestConsolidatedCtWithMultipleAccounts(t *testing.T) {
	data := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail:       true,
		ConfigureConfig:           true,
		UseConsolidatedCloudtrail: true,
		ConfigureMoreAccounts:     true,
		AwsProfile:                "main",
		Profiles: map[string]string{
			"subaccount1": "us-east-1",
			"subaccount2": "us-east-2",
		},
	})

	strippedData := strings.ReplaceAll(strings.ReplaceAll(data, "\n", ""), " ", "")
	assert.Contains(t, strippedData, "provider\"aws\"{alias=\"main\"profile=\"main\"}")
	assert.Contains(t, strippedData, "providers={aws=aws.main}")
	assert.Contains(t, strippedData, "module\"aws_config_subaccount1\"")
	assert.Contains(t, strippedData, "providers={aws=aws.subaccount1}")
	assert.Contains(t, strippedData, "provider\"aws\"{alias=\"subaccount1\"profile=\"subaccount1\"region=\"us-east-1\"}")
	assert.Contains(t, strippedData, "module\"aws_config_subaccount2\"")
	assert.Contains(t, strippedData, "providers={aws=aws.subaccount2}")
	assert.Contains(t, strippedData, "provider\"aws\"{alias=\"subaccount2\"profile=\"subaccount2\"region=\"us-east-2\"}")
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

var moduleImportCtWithConfig = `module "main_cloudtrail" {
  source                = "lacework/cloudtrail/aws"
  version               = "~> 0.1"
  iam_role_arn          = module.aws_config.iam_role_arn
  iam_role_external_id  = module.aws_config.external_id
  iam_role_name         = module.aws_config.iam_role_name
  use_existing_iam_role = true
}

`
var moduleImportCtWithoutConfig = `module "main_cloudtrail" {
  source  = "lacework/cloudtrail/aws"
  version = "~> 0.1"
}

`

var moduleImportConfig = `module "aws_config" {
  source  = "lacework/config/aws"
  version = "~> 0.1"
}

`
