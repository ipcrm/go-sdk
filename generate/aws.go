package generate

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

var (
	GenerateAwsCommandState = &GenerateAwsTfConfigurationArgs{}
)

type GenerateAwsTfConfigurationArgs struct {
	// Should we configure cloudtrail integration in LW?
	ConfigureCloudtrail bool

	// Should we configure CSPM integration in LW?
	ConfigureConfig bool

	// Supply an AWS region for where to find the cloudtrail resources
	// TODO This could be different (s3 one place, sns another)
	AwsRegion string

	// Supply an AWS Profile name for the main account, only asked if configuring multiple
	AwsProfile string

	// Use an existing trail?
	UseExistingCloudtrail bool

	// Existing S3 Bucket ARN (Required when using existing cloudtrail)
	ExistingBucketArn string

	// Use existing IAM role?
	UseExistingIamRole bool

	// Existing IAM Role ARN
	ExistingIamRoleArn string

	// Existing IAM Role Name
	ExistingIamRoleName string

	// Existing IAM Role External Id
	ExistingIamRoleExternalId string

	// Existing SNS Topic
	ExistingSnsTopicArn string

	// Consolidated Trail
	UseConsolidatedCloudtrail bool

	// Should we force destroy the bucket if it has stuff in it?
	ForceDestroyS3Bucket bool

	// For AWS Subaccounts in consolidated CT setups
	// TODO what about many ct/config integrations together?
	Profiles map[string]string

	// For aws subaccounts, a quick value to check if we are configuring multiple
	ConfigureMoreAccounts bool

	// Optional. Lacework Profile to use
	LaceworkProfile string

	// Internal CLI use
	ConfigureCloudtrailCli bool
	ConfigureConfigCli     bool
}

func NewAwsTFConfiguration(args *GenerateAwsTfConfigurationArgs) string {
	return CreateHclStringOutput(
		CombineHclBlocks(
			addRequiredProviders(),
			createAwsProviderBlock(args),
			createLaceworkProviderBlock(args),
			createConfigBlock(args),
			createCloudtrailBlock(args),
		))
}

func addRequiredProviders() *hclwrite.Block {
	return CreateRequiredProviders([]*HclRequiredProvider{{
		Name:    "lacework",
		Source:  "lacework/lacework",
		Version: "~> 0.3"}},
	)
}

func createAwsProviderBlock(args *GenerateAwsTfConfigurationArgs) []*hclwrite.Block {
	blocks := []*hclwrite.Block{}
	if args.AwsRegion != "" || args.ConfigureMoreAccounts {
		attrs := map[string]interface{}{}
		if args.AwsRegion != "" {
			attrs["region"] = args.AwsRegion
		}

		if args.ConfigureMoreAccounts {
			attrs["alias"] = "main"
			attrs["profile"] = args.AwsProfile
		}

		blocks = append(blocks, CreateProvider(&HclProvider{
			Name:       "aws",
			Attributes: attrs,
		}))
	}

	if args.ConfigureMoreAccounts {
		for profile, region := range args.Profiles {
			blocks = append(blocks, CreateProvider(&HclProvider{
				Name: "aws",
				Attributes: map[string]interface{}{
					"alias":   profile,
					"profile": profile,
					"region":  region,
				},
			}))
		}
	}

	return blocks
}

func createLaceworkProviderBlock(args *GenerateAwsTfConfigurationArgs) *hclwrite.Block {
	if args.LaceworkProfile != "" {
		return CreateProvider(&HclProvider{
			Name: "lacework",
			Attributes: map[string]interface{}{
				"profile": args.LaceworkProfile,
			},
		})
	}
	return nil
}

func createConfigBlock(args *GenerateAwsTfConfigurationArgs) []*hclwrite.Block {
	source := "lacework/config/aws"
	version := "~> 0.1"

	blocks := []*hclwrite.Block{}
	if args.ConfigureConfig {
		// Add main account
		block := &HclModule{
			Name:    "aws_config",
			Source:  source,
			Version: version,
		}

		if args.ConfigureMoreAccounts {
			block.ProviderDetails = map[string]string{
				"aws": "aws.main",
			}
		}
		blocks = append(blocks, CreateModule(block))

		// Add sub accounts
		for profile := range args.Profiles {
			blocks = append(blocks, CreateModule(&HclModule{
				Name:    fmt.Sprintf("aws_config_%s", profile),
				Source:  source,
				Version: version,
				ProviderDetails: map[string]string{
					"aws": fmt.Sprintf("aws.%s", profile),
				},
			}))
		}
	}

	return blocks
}

func createCloudtrailBlock(args *GenerateAwsTfConfigurationArgs) *hclwrite.Block {
	if args.ConfigureCloudtrail {
		data := &HclModule{
			Name:       "main_cloudtrail",
			Source:     "lacework/cloudtrail/aws",
			Version:    "~> 0.1",
			Attributes: map[string]interface{}{},
		}
		if args.ForceDestroyS3Bucket && args.UseExistingCloudtrail == false {
			data.Attributes["bucket_force_destroy"] = true
		}

		if args.UseConsolidatedCloudtrail {
			data.Attributes["consolidated_trail"] = true
		}

		if args.ExistingSnsTopicArn != "" {
			data.Attributes["use_existing_sns_topic"] = true
			data.Attributes["sns_topic_arn"] = args.ExistingSnsTopicArn
		}

		if args.UseExistingIamRole != true && args.ConfigureConfig {
			data.Attributes["use_existing_iam_role"] = true
			data.Attributes["iam_role_name"] = CreateSimpleTraversal([]string{"module", "aws_config", "iam_role_name"})
			data.Attributes["iam_role_arn"] = CreateSimpleTraversal([]string{"module", "aws_config", "iam_role_arn"})
			data.Attributes["iam_role_external_id"] = CreateSimpleTraversal([]string{"module", "aws_config", "external_id"})
		}

		if args.UseExistingIamRole {
			data.Attributes["use_existing_iam_role"] = true
			data.Attributes["iam_role_name"] = args.ExistingIamRoleName
			data.Attributes["iam_role_arn"] = args.ExistingIamRoleArn
			data.Attributes["iam_role_external_id"] = args.ExistingIamRoleExternalId
		}

		if args.UseExistingCloudtrail {
			data.Attributes["use_existing_cloudtrail"] = true
			data.Attributes["bucket_arn"] = args.ExistingBucketArn
		}

		if args.ConfigureMoreAccounts {
			data.ProviderDetails = map[string]string{
				"aws": "aws.main",
			}
		}
		return CreateModule(data)
	}

	return nil
}
