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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSimpleCSV(t *testing.T) {
	expectedCsv := strings.TrimPrefix(`
KEY,VALUE
key1,value1
key2,value2
key3,value3
`, "\n")

	csv := RenderCSV{
		headers: []string{"KEY", "VALUE"},
		data: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
				{"key3", "value3"},
	  },
	}

	assert.Equal(t,
   	csv.Render(),
		expectedCsv,
		"csv is not being formatted correctly")
}

func TestCSVHeaderCleanup(t *testing.T) {
	csv := RenderCSV{
		headers: []string{"Test header\n1", "Test header\n2"},
	}

	hasNewline := false
	for _, v := range csv.Headers() {
		if strings.Contains(v, "\n") {
			hasNewline = true
		}
	}

	assert.False(t, hasNewline, "headers are not being cleaned up properly")
}

func TestCSVDataCleanup(t *testing.T) {
	csv := RenderCSV{
		data: [][]string{{"Test data\n1"}, {"Test data\n2"}},
	}

	hasNewline := false
	for _, v := range csv.Data() {
		for _, i := range v {
			if strings.Contains(i, "\n") {
				hasNewline = true
			}
		}
	}

	assert.False(t, hasNewline, "data is not being cleaned up properly")
}

func TestTableRenderAsCSV(t *testing.T) {
  expectedCsv := strings.TrimPrefix(`
KEY,VALUE
key1,value1
key2,value2
key3,value3
`, "\n")

  table := Table{
		label: "table",
		headers: []string{"KEY", "VALUE"},
		data: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
				{"key3", "value3"},
	  },
		csvSection: "table",
  }

	assert.Equal(t,
		table.RenderAsCSV(),
		expectedCsv,
		"table is not being converted to csv properly",
  )

}
