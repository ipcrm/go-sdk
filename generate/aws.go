package generate

import (
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
	CloudTrailRegion string

	// Existing S3 Bucket ARN
	ExistingBucketArn string

	// Existing IAM Role ARN
	ExistingIamRoleArn string

	// Existing SNS Topic
	ExistingSnsTopicName string

	// Consolidated Trail
	UseConsolidatedCloudtrail bool

	// Should we force destroy the bucket if it has stuff in it?
	ForceDestroyS3Bucket bool

	// For AWS Subaccounts in consolidated CT setups
	// TODO what about many ct/config integrations together?
	Profiles []string

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
	s.CreateModuleImports()
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
	if s.Args.CloudTrailRegion != "" {
		s.Blocks = append(s.Blocks, CreateProvider(&HclProvider{
			Name: "aws",
			Attributes: map[string]interface{}{
				"region": s.Args.CloudTrailRegion,
			},
		}))
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

func (s *GenerateAwsTfConfiguration) CreateModuleImports() {
	blocks := []*hclwrite.Block{}
	if s.Args.ConfigureCloudtrail {
		data := &HclModule{
			Name:       "main_cloudtrail",
			Source:     "lacework/cloudtrail/aws",
			Version:    "~> 0.1",
			Attributes: map[string]interface{}{},
		}
		if s.Args.ForceDestroyS3Bucket {
			data.Attributes["bucket_force_destroy"] = true
		}
		blocks = append(blocks, CreateModule(data))
	}
	if s.Args.ConfigureConfig {
		blocks = append(blocks,
			CreateModule(&HclModule{
				Name:    "aws_config_main",
				Source:  "lacework/config/aws",
				Version: "~> 0.1",
			}))
	}

	s.Blocks = append(s.Blocks, blocks...)
}
