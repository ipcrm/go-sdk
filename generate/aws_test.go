package generate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerationCloudTrail(t *testing.T) {
	hcl := NewAwsTFConfiguration(&GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportCtWithoutConfig, hcl)
}

func TestGenerationConfig(t *testing.T) {
	hcl := NewAwsTFConfiguration(&GenerateAwsTfConfigurationArgs{
		ConfigureConfig: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportConfig, hcl)
}

func TestGenerationConfigAndCt(t *testing.T) {
	hcl := NewAwsTFConfiguration(&GenerateAwsTfConfigurationArgs{
		ConfigureConfig: true, ConfigureCloudtrail: true,
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+moduleImportConfig+moduleImportCtWithConfig, hcl)
}

func TestGenerationWithProviderRegion(t *testing.T) {
	hcl := NewAwsTFConfiguration(&GenerateAwsTfConfigurationArgs{
		AwsRegion: "us-east-2",
	})
	assert.NotNil(t, hcl)
	assert.Equal(t, requiredProviders+awsProvider, hcl)
}

func TestGenerationCloudtrailForceDestroyS3(t *testing.T) {
	data := createCloudtrailBlock(
		&GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:  true,
			ForceDestroyS3Bucket: true,
		},
	)
	assert.Equal(t,
		"bucket_force_destroy=true\n",
		string(data.Body().GetAttribute("bucket_force_destroy").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailConsolidatedTrail(t *testing.T) {
	data := createCloudtrailBlock(
		&GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:       true,
			UseConsolidatedCloudtrail: true,
		},
	)
	assert.Equal(t,
		"consolidated_trail=true\n",
		string(data.Body().GetAttribute("consolidated_trail").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailExistingSns(t *testing.T) {
	existingSnsTopicArn := "arn:aws:sns:::foo"
	data := createCloudtrailBlock(
		&GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail: true,
			ExistingSnsTopicArn: existingSnsTopicArn,
		},
	)
	assert.Equal(t,
		fmt.Sprintf("sns_topic_arn=\"%s\"\n", existingSnsTopicArn),
		string(data.Body().GetAttribute("sns_topic_arn").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		"use_existing_sns_topic=true\n",
		string(data.Body().GetAttribute("use_existing_sns_topic").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailExistingBucket(t *testing.T) {
	existingBucketArn := "arn:aws:s3:::test-bucket-12345"
	data := createCloudtrailBlock(
		&GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:   true,
			UseExistingCloudtrail: true,
			ExistingBucketArn:     existingBucketArn,
		},
	)
	assert.Equal(t,
		"use_existing_cloudtrail=true\n",
		string(data.Body().GetAttribute("use_existing_cloudtrail").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("bucket_arn=\"%s\"\n", existingBucketArn),
		string(data.Body().GetAttribute("bucket_arn").BuildTokens(nil).Bytes()))
}

func TestGenerationCloudtrailExistingRole(t *testing.T) {
	iamRoleArn := "arn:aws:iam::123456789012:role/test-role"
	iamRoleName := "test-role"
	extId := "1234567890123456"

	data := createCloudtrailBlock(
		&GenerateAwsTfConfigurationArgs{
			ConfigureCloudtrail:       true,
			UseExistingIamRole:        true,
			ExistingIamRoleArn:        iamRoleArn,
			ExistingIamRoleName:       iamRoleName,
			ExistingIamRoleExternalId: extId,
		},
	)
	assert.Equal(t,
		"use_existing_iam_role=true\n",
		string(data.Body().GetAttribute("use_existing_iam_role").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("iam_role_name=\"%s\"\n", iamRoleName),
		string(data.Body().GetAttribute("iam_role_name").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("iam_role_arn=\"%s\"\n", iamRoleArn),
		string(data.Body().GetAttribute("iam_role_arn").BuildTokens(nil).Bytes()))
	assert.Equal(t,
		fmt.Sprintf("iam_role_external_id=\"%s\"\n", extId),
		string(data.Body().GetAttribute("iam_role_external_id").BuildTokens(nil).Bytes()))
}

func TestConsolidatedCtWithMultipleAccounts(t *testing.T) {
	data := NewAwsTFConfiguration(&GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail:       true,
		ConfigureConfig:           true,
		UseConsolidatedCloudtrail: true,
		ConfigureSubAccounts:      true,
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
