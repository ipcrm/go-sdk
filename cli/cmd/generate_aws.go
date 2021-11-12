package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/lacework/go-sdk/generate"
	"github.com/pkg/errors"
)

func promptAwsCtQuestions(config *generate.GenerateAwsTfConfigurationArgs) error {
	// Only ask these questions if configure cloudtrail is true
	if err := SurveyMultipleQuestionWithValidation([]SurveyQuestionWithValidationArgs{
		{
			Prompt:   &survey.Confirm{Message: "Use consolidated Cloudtrail?", Default: config.UseConsolidatedCloudtrail},
			Response: &config.UseConsolidatedCloudtrail,
		},
		{
			Prompt:   &survey.Confirm{Message: "Use existing cloudtrail?", Default: config.UseExistingCloudtrail},
			Response: &config.UseExistingCloudtrail,
		},
		{
			Prompt: &survey.Input{
				Message: "Specify an existing bucket ARN used for Cloudtrail logs:",
				Default: config.ExistingBucketArn,
			},
			Checks:   []*bool{&config.UseExistingCloudtrail},
			Required: true,
			Response: &config.ExistingBucketArn,
		},
	}, config.ConfigureCloudtrail); err != nil {
		return err
	}

	// If a new bucket is to be created; should the force destroy bit be set?
	newBucket := config.ExistingBucketArn == ""
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Prompt: &survey.Confirm{
			Message: "Should the new S3 bucket have force destroy enabled?",
			Default: config.ForceDestroyS3Bucket},
		Response: &config.ForceDestroyS3Bucket,
		Checks:   []*bool{&config.ConfigureCloudtrail, &newBucket}}); err != nil {
		return err
	}

	return nil
}

func promptAwsExistingIamQuestions(config *generate.GenerateAwsTfConfigurationArgs) error {
	if err := SurveyMultipleQuestionWithValidation([]SurveyQuestionWithValidationArgs{
		{
			Prompt: &survey.Input{
				Message: "Specify an existing IAM role name for Cloudtrail access",
				Default: config.ExistingIamRoleName},
			Response: &config.ExistingIamRoleName,
			Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		},
		{
			Prompt: &survey.Input{
				Message: "Specify an existing IAM role ARN for Cloudtrail access",
				Default: config.ExistingIamRoleArn,
			},
			Response: &config.ExistingIamRoleArn,
			Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		},
		{
			Prompt: &survey.Input{
				Message: "Specify the external ID to be used with the existing IAM role",
				Default: config.ExistingIamRoleExternalId,
			},
			Response: &config.ExistingIamRoleExternalId,
			Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		}}, config.UseExistingIamRole); err != nil {
		return err
	}

	return nil
}

func promptAwsAdditionalAccountQuestions(config *generate.GenerateAwsTfConfigurationArgs) error {
	type accountAnswers struct {
		AccountProfileName   string `survey:"accountProfileName"`
		AccountProfileRegion string `survey:"accountProfileRegion"`
	}
	accountQuestions := []*survey.Question{
		{
			Name:     "accountProfileName",
			Prompt:   &survey.Input{Message: "Supply the profile name for this additional AWS account"},
			Validate: survey.Required,
		},
		{
			Name:     "accountProfileRegion",
			Prompt:   &survey.Input{Message: "What region should be used for this account?"},
			Validate: survey.Required,
		},
	}

	// For each added account, collect it's profile name and the region that should be used
	accountDetails := map[string]string{}
	askAgain := true

	// Determine the profile for the main account
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Prompt: &survey.Input{
			Message: "What is the AWS profile name for the main account?", // TODO @ipcrm Make this prompt better
			Default: config.AwsProfile,
			Help:    "This is the main account where your cloudtrail resources are created",
		},
		Response: &config.AwsProfile,
		Required: true}); err != nil {
		return nil
	}

	// For each account to add, collect the aws profile and region to use
	for askAgain {
		answers := accountAnswers{}
		err := survey.Ask(accountQuestions, &answers)
		if err != nil {
			return err
		}
		accountDetails[answers.AccountProfileName] = answers.AccountProfileRegion

		err = survey.AskOne(
			&survey.Confirm{Message: "Add another AWS account?"},
			&askAgain,
		)
		if err != nil {
			return err
		}
	}
	config.Profiles = accountDetails

	return nil
}

func setValsFromCliInput(config *generate.GenerateAwsTfConfigurationArgs) {
	// If a bucket arn was supplied, we are using an existing ct
	if config.ExistingBucketArn != "" {
		config.UseExistingCloudtrail = true
	}

	// If any of these were set in the command line args we need to set useexistingiamrole to true and prompt for what we
	// are missing
	if config.ExistingIamRoleArn != "" ||
		config.ExistingIamRoleName != "" ||
		config.ExistingIamRoleExternalId != "" {
		config.UseExistingIamRole = true
	}

}

func validateInputCombinations(config *generate.GenerateAwsTfConfigurationArgs) error {
	// Validate that at least region was set
	if config.ConfigureCloudtrail && config.AwsRegion == "" {
		return errors.New("AWS Region must be set when configuring Cloudtrail!")
	}

	// Validate if using an existing cloudtrail the bucket was provided
	if config.UseExistingCloudtrail && config.ExistingBucketArn == "" {
		return errors.New("Must supply bucket ARN when using an existing cloudtrail!")
	}

	// Validate required values got set one way or another
	// If this was run non-interactive and parts of the data are missing, error out
	if config.UseExistingIamRole {
		if config.ExistingIamRoleArn == "" ||
			config.ExistingIamRoleName == "" ||
			config.ExistingIamRoleExternalId == "" {
			return errors.New("When using an existing IAM role, the existing role ARN, Name, and External ID all must be set!")
		}
	}

	return nil
}

func askAdvancedOptions(config *generate.GenerateAwsTfConfigurationArgs) error {
	// Construction of this slice is a bit strange at first look, but the reason for that is because we have to do string
	// validation to know which option was selected due to how survey works // TODO @ipcrm is doing this by index easier?
	// TODO This needs to be selective based on what was supplied
	askCloudTrailOptions := "Additional Cloudtrail Options"
	askIamRoleOptions := "Configure Lacework integration with an existing IAM role"
	askAdditionalAwsAccountsOptions := "Add Additional AWS Accounts to Lacework"
	askCustomizeOutputLocationOptions := "Customize Output Location"
	done := "Done"
	options := []string{askCloudTrailOptions, askIamRoleOptions, askAdditionalAwsAccountsOptions, askCustomizeOutputLocationOptions, done}
	answer := ""

	for answer != "Done" {
		if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
			Prompt: &survey.Select{
				Message: "Which options would you like to enable?",
				Options: options,
			},
			Response: &answer,
		}); err != nil {
			return err
		}

		switch answer {
		case askCloudTrailOptions:
			if err := promptAwsCtQuestions(config); err != nil {
				return err
			}
		case askIamRoleOptions:
			config.UseExistingIamRole = true
			if err := promptAwsExistingIamQuestions(config); err != nil {
				return nil
			}

		case askAdditionalAwsAccountsOptions:
			config.ConfigureSubAccounts = true
			if err := promptAwsAdditionalAccountQuestions(config); err != nil {
				return err
			}
		}

		// Re-prompt
		innerAskAgain := true
		if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
			Checks:   []*bool{&innerAskAgain},
			Prompt:   &survey.Confirm{Message: "Configure another advanced integration option", Default: false},
			Response: &innerAskAgain,
		}); err != nil {
			return err
		}

		// TODO @ipcrm this needs improved
		if !innerAskAgain {
			answer = "Done"
		}
	}

	return nil
}

func promptAwsGenerate(config *generate.GenerateAwsTfConfigurationArgs) error {
	// Set vals that were passed in, where required
	setValsFromCliInput(config)

	// This are the core questions that should be asked.  Region required for provider block
	if err := SurveyMultipleQuestionWithValidation(
		[]SurveyQuestionWithValidationArgs{
			{
				Prompt:   &survey.Confirm{Message: "Enable Config Integration?", Default: config.ConfigureConfig},
				Response: &config.ConfigureConfig,
			},
			{
				Prompt:   &survey.Confirm{Message: "Enable Cloudtrail Integration?", Default: config.ConfigureCloudtrail},
				Response: &config.ConfigureCloudtrail,
			},
			{
				Checks: []*bool{&config.ConfigureConfig, &config.ConfigureCloudtrail},
				Prompt: &survey.Input{
					Message: "Specify the AWS region Cloudtrail, SNS, and S3 resources should use:",
					Default: config.AwsRegion,
				},
				Response: &config.AwsRegion,
				Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)},
			},
		}); err != nil {
		return err
	}

	// Validate one of config or cloudtrail was enabled; otherwise error out
	if !config.ConfigureCloudtrail && !config.ConfigureConfig {
		return errors.New("Must enable cloudtrail or config!")
	}

	// Find out if the customer wants to specify more advanced features
	askAdvanced := false
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Prompt:   &survey.Confirm{Message: "Configure advanced integration options?", Default: askAdvanced},
		Response: &askAdvanced,
	}); err != nil {
		return err
	}

	// Keep prompting for advanced options until the say done
	if askAdvanced {
		if err := askAdvancedOptions(config); err != nil {
			return err
		}
	}

	// Validate the must haves for input combinations
	if err := validateInputCombinations(config); err != nil {
		return err
	}

	return nil
}
