package cmd

import (
	"testing"

	"github.com/lacework/go-sdk/generate"
	"github.com/stretchr/testify/assert"
)

func toggleNonInteractive() {
	cli.nonInteractive = !cli.nonInteractive
}

func TestGenerateMostBasicArgs(t *testing.T) {
	toggleNonInteractive()
	defer toggleNonInteractive()

	data := generate.GenerateAwsTfConfiguration{}
	data.ConfigureCloudtrailCli = true
	data.ConfigureConfigCli = true
	data.AwsRegion = "us-east-2"
	err := promptAwsGenerate(&data)

	assert.Nil(t, err)
	assert.True(t, data.ConfigureCloudtrail)
	assert.True(t, data.ConfigureConfig)
}

func TestMissingExistingIamRoleParams(t *testing.T) {
	toggleNonInteractive()
	defer toggleNonInteractive()

	data := generate.GenerateAwsTfConfiguration{}
	data.ConfigureCloudtrailCli = true
	data.ExistingIamRoleArn = "blue"
	err := promptAwsGenerate(&data)

	// This should get set automatically if any existing role details are set
	assert.True(t, data.UseExistingIamRole)

	// This should error out, we need to set all existing iam role details
	assert.Error(t, err)
}

func TestMissingExistingCloudtrailParams(t *testing.T) {
	toggleNonInteractive()
	defer toggleNonInteractive()

	data := generate.GenerateAwsTfConfiguration{}
	data.ConfigureCloudtrail = true
	data.UseExistingCloudtrail = true
	data.AwsRegion = "us-east-2"

	err := promptAwsGenerate(&data)
	assert.Error(t, err)
	assert.Equal(t, "Must supply bucket ARN when using an existing cloudtrail!", err.Error())
}
