//
// Author:: Salim Afiune Maya (<afiune@lacework.net>)
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

package lwupdater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	// GithubOrganization is the default Github organization where
	// Lacework stores their open source projects
	GithubOrganization = "lacework"

	// DisableEnv controls the overall check for updates behavior, when
	// this environment variable is set, we do not check for updates
	DisableEnv = "LW_UPDATES_DISABLE"
)

type info struct {
	Project  string
	Version  string
	Latest   string
	Outdated bool
}

// Check verifies if the a project is outdated based of the current version
func Check(project, current string) (*info, error) {
	if disabled := os.Getenv(DisableEnv); disabled != "" {
		return new(info), nil
	}

	release, err := getGitRelease(project, "latest")
	if err != nil {
		return new(info), err
	}

	return &info{
		Project:  project,
		Version:  current,
		Latest:   release.TagName,
		Outdated: current != release.TagName,
	}, nil
}

// getGitRelease uses the git API to fetch the release information of a project.
// This function could hit request rate limits wich are roughly 60 every 30m, to
// check your current rate limits run: curl https://api.github.com/rate_limit
//
// TODO @afiune implement a cache mechanism
func getGitRelease(project, version string) (*gitReleaseResponse, error) {
	if project == "" {
		return nil, errors.New("specify a valid project")
	}
	if version == "" {
		version = "latest"
	}

	var (
		c = http.Client{}
		u = url.URL{
			Scheme: "https",
			Host:   "api.github.com",
			Path: fmt.Sprintf(
				"/repos/%s/%s/releases/latest",
				GithubOrganization, project,
			),
		}
	)
	if version != "latest" {
		u.Path = fmt.Sprintf("/repos/%s/%s/releases/tags/%s",
			GithubOrganization, project, version)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the user agent since it is required
	// https://developer.github.com/v3/#user-agent-required
	req.Header.Set("User-Agent", "lacework-updater")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if c := resp.StatusCode; c >= 200 && c <= 299 {
		var gitRelRes gitReleaseResponse
		if err := json.NewDecoder(resp.Body).Decode(&gitRelRes); err != nil {
			return nil, err
		}

		return &gitRelRes, nil
	}

	// not a successful response, throw an error
	return nil, errors.New(resp.Status)
}

type gitReleaseResponse struct {
	ID              int32     `json:"id"`
	Url             string    `json:"url"`
	HtmlUrl         string    `json:"html_url"`
	AssetsUrl       string    `json:"assets_url"`
	UploadUrl       string    `json:"upload_url"`
	TarballUrl      string    `json:"tarball_url"`
	ZipballUrl      string    `json:"zipball_url"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Body            string    `json:"body"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
}
