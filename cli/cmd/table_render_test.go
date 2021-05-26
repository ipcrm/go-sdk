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

package cmd

import (
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/assert"
)

func TestRenderSimpleTable(t *testing.T) {
	expectedTable := strings.TrimPrefix(`
  KEY    VALUE   
-------+---------
  key1   value1  
  key2   value2  
  key3   value3  
`, "\n")

	assert.Equal(t,
		renderSimpleTable(
			[]string{"KEY", "VALUE"},
			[][]string{
				[]string{"key1", "value1"},
				[]string{"key2", "value2"},
				[]string{"key3", "value3"},
		}),
		expectedTable,
		"tables are not being formatted correctly")
}

func TestRenderSimpleTableLongDescriptions(t *testing.T) {
	expectedTable := strings.TrimPrefix(`
  ID            DESCRIPTION            
-----+---------------------------------
  1    This is a long long very        
       long description that will be   
       splitted into multiple lines    
  2    No a very long description      
`, "\n")

	assert.Equal(t,
		renderSimpleTable(
			[]string{"ID", "Description"},
			[][]string{
				[]string{"1", "This is a long long very long description that will be splitted into multiple lines"},
				[]string{"2", "No a very long description"},
			}),
		expectedTable,
		"tables are not being formatted correctly")
}

func TestSimpleTable(t *testing.T) {
	expectedTable := strings.TrimPrefix(`
  KEY    VALUE   
-------+---------
  key1   value1  
  key2   value2  
  key3   value3  
`, "\n")

	assert.Equal(t,
		renderSimpleTable(
			[]string{"KEY", "VALUE"},
			[][]string{
							{"key1", "value1"},
							{"key2", "value2"},
							{"key3", "value3"},
			},
		),
		expectedTable,
		"tables are not being formatted correctly")
}

func TestSimpleTableLongDescriptions(t *testing.T) {
	expectedTable := strings.TrimPrefix(`
  ID            DESCRIPTION            
-----+---------------------------------
  1    This is a long long very        
       long description that will be   
       splitted into multiple lines    
  2    No a very long description      
`, "\n")

	assert.Equal(t,
		renderSimpleTable(
			[]string{"ID", "Description"},
			[][]string{
				{"1", "This is a long long very long description that will be splitted into multiple lines"},
				{"2", "No a very long description"},
			},
		),
		expectedTable,
		"tables are not being formatted correctly")
}

func TestCustomTable(t *testing.T) {
	detailsTable := [][]string{
		{"KEY1", "VALUE1"},
		{"KEY2", "VALUE2"},
		{"KEY3", "VALUE3"},
	}
	summaryTable := [][]string{
		{"Severity1", "1"},
		{"Secerity2", "2"},
		{"Secerity3", "0"},
	}
	expectedTable := strings.TrimPrefix(`
   REPORT DETAILS       RECOMMENDATIONS     
-------------------+------------------------
    KEY1  VALUE1       SEVERITY    COUNT    
    KEY2  VALUE2     ------------+--------  
    KEY3  VALUE3       Severity1       1    
                       Secerity2       2    
                       Secerity3       0    
                                            
`, "\n")

	assert.Equal(t,
		renderCustomTable(
			"overview",
		  []string{"Report Details", "Recommendations"},
			nil,
			[]*Table{
				NewTable(
					"details",
					nil,
					detailsTable,
					nil,
					tableFunc(func(t *tablewriter.Table) {
						t.SetBorder(false)
						t.SetColumnSeparator("")
						t.SetAlignment(tablewriter.ALIGN_LEFT)
					}),
				),
				NewTable(
					"summary",
					[]string{"Severity", "Count"},
					summaryTable,
					nil,
					tableFunc(func(t *tablewriter.Table) {
						t.SetBorder(false)
						t.SetColumnSeparator(" ")
					}),
				),
			},
			tableFunc(func(t *tablewriter.Table) {
				t.SetBorder(false)
				t.SetAutoWrapText(false)
				t.SetColumnSeparator(" ")
			}),
		),
		expectedTable,
		"tables are not being formatted correctly")
}
func TestThisTable(t *testing.T) {
	expectedTable := strings.TrimPrefix(`
ID,Recommendation,Status,Severity,Service,Affected,Assessed
LW_S3_18,...global 'Get' permission.,Compliant,Critical,aws:s3,0,0
LW_S3_2,"...not grant 'Everyone' WRITE permission [create, overwrite, and delete S3 objects]",Compliant,Critical,aws:s3,0,0
LW_S3_3,..'Everyone' READ_ACP permission [read bucket ACL],Compliant,Critical,aws:s3,0,0
`, "\n")
	data := [][]string{
		{"LW_S3_18","...global 'Get' permission.","Compliant","Critical","aws:s3","0","0"},
	  {"LW_S3_2","...not grant 'Everyone' WRITE permission [create, overwrite, and delete S3 objects]","Compliant","Critical","aws:s3","0","0"},
	  {"LW_S3_3","..'Everyone' READ_ACP permission [read bucket ACL]","Compliant","Critical","aws:s3","0","0"}}

  table	:= NewTable(
			"recommendations",
			[]string{"ID", "Recommendation", "Status", "Severity", "Service", "Affected", "Assessed"},
			data,
			nil,
			tableFunc(func(t *tablewriter.Table) {
				t.SetBorder(false)
				t.SetRowLine(true)
				t.SetColumnSeparator(" ")
			}),
  )

	table.renderAsCsv = true
	table.csvSection = "recommendations"
	assert.Equal(t,
	  table.Render(),
		expectedTable,
	)
}

