//
// Author:: Matt Cadorette (<matthew.cadorette@lacework.net>)
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

package cmd

import (
	"encoding/csv"
	"fmt"
	"strings"
)

type RenderCSV struct {
	headers []string
	data    [][]string
}

// Return headers in current RenderCSV, stripping any newlines
func (r *RenderCSV) Headers() []string {
	var newHeaders []string
	for _, h := range r.headers {
		newHeaders = append(newHeaders, strings.Replace(h, "\n", "", -1))
	}
	return newHeaders
}

// Return data in current RenderCSV, stripping any newlines
func (r *RenderCSV) Data() [][]string {
	var newData [][]string
	for _, d := range r.data {
		var newInnerData []string
		for _, id := range d {
			newInnerData = append(newInnerData, strings.Replace(id, "\n", "", -1))
		}
		newData = append(newData, newInnerData)
	}
	return newData
}

// Used to produce CSV output
func (r *RenderCSV) Render() string {
	csvOut := &strings.Builder{}
	csvRaw := &strings.Builder{}
	csv := csv.NewWriter(csvOut)

	if len(r.Headers()) > 0 {
		csv.Write(r.Headers())
	}

	for _, record := range r.Data() {
		if err := csv.Write(record); err != nil {
			fmt.Printf("Failed to build csv output, got error: %s", err.Error())
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	csv.Flush()
	return csvOut.String() + csvRaw.String()
}

// Helper to convert table to CSV format using RenderCSV
func renderTableAsCSV(headers []string, data [][]string) string {
	r := new(RenderCSV)
	r.headers = headers
	r.data = data
	return r.Render()
}
