package generate

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type GenerateAwsTfConfigurationArgs struct {
	// Supplied values

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
}

type GenerateAwsTfConfiguration struct {
	Args *GenerateAwsTfConfigurationArgs

	Blocks []*hclwrite.Block
}

func NewAwsTFConfiguration(args *GenerateAwsTfConfigurationArgs) string {
	s := &GenerateAwsTfConfiguration{
		Args: args,
	}
	s.AddRequiredProviders()
	s.CreateAwsProviderBlock()
	s.CreateLaceworkProviderBlock()
	s.CreateConfigBlock()
	s.CreateCloudtrailBlock()
	return CreateHclStringOutput(s.Blocks)
}

func (s *GenerateAwsTfConfiguration) AddRequiredProviders() {
	s.Blocks = append(s.Blocks,
		CreateRequiredProviders([]*HclRequiredProvider{{
			Name:    "lacework",
			Source:  "lacework/lacework",
			Version: "~> 0.3"}},
		),
	)
}

func (s *GenerateAwsTfConfiguration) CreateAwsProviderBlock() {
	if s.Args.AwsRegion != "" || s.Args.ConfigureMoreAccounts {
		attrs := map[string]interface{}{}
		if s.Args.AwsRegion != "" {
			attrs["region"] = s.Args.AwsRegion
		}

		if s.Args.ConfigureMoreAccounts {
			attrs["alias"] = "main"
			attrs["profile"] = s.Args.AwsProfile
		}
		s.Blocks = append(s.Blocks, CreateProvider(&HclProvider{
			Name:       "aws",
			Attributes: attrs,
		}))
	}

	if s.Args.ConfigureMoreAccounts {
		for profile, region := range s.Args.Profiles {
			s.Blocks = append(s.Blocks, CreateProvider(&HclProvider{
				Name: "aws",
				Attributes: map[string]interface{}{
					"alias":   profile,
					"profile": profile,
					"region":  region,
				},
			}))
		}
	}
}

func (s *GenerateAwsTfConfiguration) CreateLaceworkProviderBlock() {
	if s.Args.LaceworkProfile != "" {
		s.Blocks = append(s.Blocks, CreateProvider(&HclProvider{
			Name: "lacework",
			Attributes: map[string]interface{}{
				"profile": s.Args.LaceworkProfile,
			},
		}))
	}
}

func (s *GenerateAwsTfConfiguration) CreateConfigBlock() {
	source := "lacework/config/aws"
	version := "~> 0.1"

	if s.Args.ConfigureConfig {
		block := &HclModule{
			Name:    "aws_config",
			Source:  source,
			Version: version,
		}

		if s.Args.ConfigureMoreAccounts {
			block.ProviderDetails = map[string]string{
				"aws": "aws.main",
			}
		}
		s.Blocks = append(s.Blocks, CreateModule(block))

		for profile := range s.Args.Profiles {
			s.Blocks = append(s.Blocks, CreateModule(&HclModule{
				Name:    fmt.Sprintf("aws_config_%s", profile),
				Source:  source,
				Version: version,
				ProviderDetails: map[string]string{
					"aws": fmt.Sprintf("aws.%s", profile),
				},
			}))
		}
	}
}

func (s *GenerateAwsTfConfiguration) CreateCloudtrailBlock() {
	if s.Args.ConfigureCloudtrail {
		data := &HclModule{
			Name:       "main_cloudtrail",
			Source:     "lacework/cloudtrail/aws",
			Version:    "~> 0.1",
			Attributes: map[string]interface{}{},
		}
		if s.Args.ForceDestroyS3Bucket && s.Args.UseExistingCloudtrail == false {
			data.Attributes["bucket_force_destroy"] = true
		}

		if s.Args.UseConsolidatedCloudtrail {
			data.Attributes["consolidated_trail"] = true
		}

		if s.Args.ExistingSnsTopicArn != "" {
			data.Attributes["use_existing_sns_topic"] = true
			data.Attributes["sns_topic_arn"] = s.Args.ExistingSnsTopicArn
		}

		if s.Args.UseExistingIamRole != true && s.Args.ConfigureConfig {
			data.Attributes["use_existing_iam_role"] = true
			data.Attributes["iam_role_name"] = CreateSimpleTraversal([]string{"module", "aws_config", "iam_role_name"})
			data.Attributes["iam_role_arn"] = CreateSimpleTraversal([]string{"module", "aws_config", "iam_role_arn"})
			data.Attributes["iam_role_external_id"] = CreateSimpleTraversal([]string{"module", "aws_config", "external_id"})
		}

		if s.Args.UseExistingIamRole {
			data.Attributes["use_existing_iam_role"] = true
			data.Attributes["iam_role_name"] = s.Args.ExistingIamRoleName
			data.Attributes["iam_role_arn"] = s.Args.ExistingIamRoleArn
			data.Attributes["iam_role_external_id"] = s.Args.ExistingIamRoleExternalId
		}

		if s.Args.UseExistingCloudtrail {
			data.Attributes["use_existing_cloudtrail"] = true
			data.Attributes["bucket_arn"] = s.Args.ExistingBucketArn
		}

		if s.Args.ConfigureMoreAccounts {
			data.ProviderDetails = map[string]string{
				"aws": "aws.main",
			}
		}
		s.Blocks = append(s.Blocks, CreateModule(data))
	}
}
