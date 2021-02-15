//
// Author:: Darren Murray (<darren.murray@lacework.net>)
// Copyright:: Copyright 2020, Lacework Inc.
// License:: Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lacework/go-sdk/api"
	"github.com/lacework/go-sdk/internal/intgguid"
	"github.com/lacework/go-sdk/internal/lacework"
)

func TestIntegrationsNewMicrosoftTeamsAlertChannel(t *testing.T) {
	subject := api.NewMicrosoftTeamsAlertChannel("integration_name",
		api.MicrosoftTeamsChannelData{
			WebhookURL: "https://outlook.office.com/webhook/api-token",
		},
	)
	assert.Equal(t, api.MicrosoftTeamsChannelIntegration.String(), subject.Type)
}

func TestIntegrationsCreateMicrosoftTeamsAlertChannel(t *testing.T) {
	var (
		intgGUID   = intgguid.New()
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI("external/integrations", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "CreateMicrosoftTeamsAlertChannel should be a POST method")

		if assert.NotNil(t, r.Body) {
			body := httpBodySniffer(r)
			assert.Contains(t, body, "integration_name", "integration name is missing")
			assert.Contains(t, body, "MICROSOFT_TEAMS", "wrong integration type")
			assert.Contains(t, body, "https://outlook.office.com/webhook/api-token", "wrong microsoft Teams url")
			assert.Contains(t, body, "ENABLED\":1", "integration is not enabled")
		}

		fmt.Fprintf(w, microsoftTeamsChannelIntegrationJsonResponse(intgGUID))
	})
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	data := api.NewMicrosoftTeamsAlertChannel("integration_name",
		api.MicrosoftTeamsChannelData{
			WebhookURL: "https://outlook.office.com/webhook/api-token",
		},
	)
	assert.Equal(t, "integration_name", data.Name, "MicrosoftTeams integration name mismatch")
	assert.Equal(t, "MICROSOFT_TEAMS", data.Type, "a new Microsoft Teams integration should match its type")
	assert.Equal(t, 1, data.Enabled, "a new MicrosoftTeams integration should be enabled")

	response, err := c.Integrations.CreateMicrosoftTeamsAlertChannel(data)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Ok)
	if assert.Equal(t, 1, len(response.Data)) {
		resData := response.Data[0]
		assert.Equal(t, intgGUID, resData.IntgGuid)
		assert.Equal(t, "integration_name", resData.Name)
		assert.True(t, resData.State.Ok)
		assert.Equal(t, "https://outlook.office.com/webhook/api-token", resData.Data.WebhookURL)
	}
}

func TestIntegrationsGetMicrosoftTeamsAlertChannel(t *testing.T) {
	var (
		intgGUID   = intgguid.New()
		apiPath    = fmt.Sprintf("external/integrations/%s", intgGUID)
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI(apiPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "GetMicrosoftTeamsAlertChannel should be a GET method")
		fmt.Fprintf(w, microsoftTeamsChannelIntegrationJsonResponse(intgGUID))
	})
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	response, err := c.Integrations.GetMicrosoftTeamsAlertChannel(intgGUID)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Ok)
	if assert.Equal(t, 1, len(response.Data)) {
		resData := response.Data[0]
		assert.Equal(t, intgGUID, resData.IntgGuid)
		assert.Equal(t, "integration_name", resData.Name)
		assert.True(t, resData.State.Ok)
		assert.Equal(t, "https://outlook.office.com/webhook/api-token", resData.Data.WebhookURL)
	}
}

func TestIntegrationsUpdateMicrosoftTeamsAlertChannel(t *testing.T) {
	var (
		intgGUID   = intgguid.New()
		apiPath    = fmt.Sprintf("external/integrations/%s", intgGUID)
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI(apiPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method, "UpdateMicrosoftTeamsAlertChannel should be a PATCH method")

		if assert.NotNil(t, r.Body) {
			body := httpBodySniffer(r)
			assert.Contains(t, body, intgGUID, "INTG_GUID missing")
			assert.Contains(t, body, "integration_name", "integration name is missing")
			assert.Contains(t, body, "MICROSOFT_TEAMS", "wrong integration type")
			assert.Contains(t, body, "https://outlook.office.com/webhook/api-token", "wrong microsoftTeams url")
			assert.Contains(t, body, "ENABLED\":1", "integration is not enabled")
		}

		fmt.Fprintf(w, microsoftTeamsChannelIntegrationJsonResponse(intgGUID))
	})
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	data := api.NewMicrosoftTeamsAlertChannel("integration_name",
		api.MicrosoftTeamsChannelData{
			WebhookURL: "https://outlook.office.com/webhook/api-token",
		},
	)
	assert.Equal(t, "integration_name", data.Name, "MicrosoftTeams integration name mismatch")
	assert.Equal(t, "MICROSOFT_TEAMS", data.Type, "a new MicrosoftTeams integration should match its type")
	assert.Equal(t, 1, data.Enabled, "a new MicrosoftTeams integration should be enabled")
	data.IntgGuid = intgGUID

	response, err := c.Integrations.UpdateMicrosoftTeamsAlertChannel(data)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "SUCCESS", response.Message)
	assert.Equal(t, 1, len(response.Data))
	assert.Equal(t, intgGUID, response.Data[0].IntgGuid)
}

func TestIntegrationsListMicrosoftTeamsAlertChannel(t *testing.T) {
	var (
		intgGUIDs  = []string{intgguid.New(), intgguid.New(), intgguid.New()}
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI("external/integrations/type/MICROSOFT_TEAMS",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method, "ListMicrosoftTeamsAlertChannel should be a GET method")
			fmt.Fprintf(w, microsoftTeamsChanMultiIntegrationJsonResponse(intgGUIDs))
		},
	)
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	response, err := c.Integrations.ListMicrosoftTeamsAlertChannel()
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Ok)
	assert.Equal(t, len(intgGUIDs), len(response.Data))
	for _, d := range response.Data {
		assert.Contains(t, intgGUIDs, d.IntgGuid)
	}
}

func microsoftTeamsChannelIntegrationJsonResponse(intgGUID string) string {
	return `
{
  "data": [` + singleMicrosoftTeamsChanIntegration(intgGUID) + `],
  "ok": true,
  "message": "SUCCESS"
}
`
}

func microsoftTeamsChanMultiIntegrationJsonResponse(guids []string) string {
	integrations := []string{}
	for _, guid := range guids {
		integrations = append(integrations, singleMicrosoftTeamsChanIntegration(guid))
	}
	return `
{
"data": [` + strings.Join(integrations, ", ") + `],
"ok": true,
"message": "SUCCESS"
}
`
}

func singleMicrosoftTeamsChanIntegration(id string) string {
	return `
{
  "INTG_GUID": "` + id + `",
  "CREATED_OR_UPDATED_BY": "user@email.com",
  "CREATED_OR_UPDATED_TIME": "2020-Jul-16 19:59:22 UTC",
  "DATA": {
    "TEAMS_URL": "https://outlook.office.com/webhook/api-token"
  },
  "ENABLED": 1,
  "IS_ORG": 0,
  "NAME": "integration_name",
  "STATE": {
    "lastSuccessfulTime": "2020-Jul-16 18:26:54 UTC",
    "lastUpdatedTime": "2020-Jul-16 18:26:54 UTC",
    "ok": true
  },
  "TYPE": "MICROSOFT_TEAMS",
  "TYPE_NAME": "MICROSOFT_TEAMS"
}
`
}
