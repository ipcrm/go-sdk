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

	"github.com/olekukonko/tablewriter"
)

type Table struct {
	// Headers for each table field
	headers     []string

	// Data to be included in this table
	data        [][]string

	// If this table is wrapping other tables, supply those other tables here.
	// IMPORTANT: The headers for this table are auto-assigned to each inner table based on postion.
	// Example: 2 hears, 2 inner tables.  Inner table 1 is shown under header 1.
	innerTables []Table

	// Supply table options to customize behavior.  See tablewriter interface for details.
	opts        []tableOption

	// Future use; you SHOULD populate this with a human readable, lowercase, hyphenated text
	label       string

	// Future use; should this table output be rendered as a CSV
	renderAsCsv bool
}

// Apply any custom options
func (t *Table) ApplyOpts(tbl *tablewriter.Table) {
	for _, opt := range t.opts {
		opt.apply(tbl)
	}
}

// Testing
func (t *Table) Render() string {
	// TODO
	// validate only data or innerTables supplied
	var returnString string

	if t.renderAsCsv {
		return t.RenderAsCSV()
	}

	if len(t.innerTables) == 0 {
		var (
			tblBldr = &strings.Builder{}
			tbl     = tablewriter.NewWriter(tblBldr)
		)
		tbl.SetHeader(t.headers)

		// Custom table
		if len(t.opts) > 0 {
			t.ApplyOpts(tbl)
		} else {
			// Simple table
			tbl.SetRowLine(false)
			tbl.SetBorder(false)
			tbl.SetAutoWrapText(true)
			tbl.SetAlignment(tablewriter.ALIGN_LEFT)
			tbl.SetColumnSeparator(" ")
		}

		tbl.AppendBulk(t.data)
		tbl.Render()
		returnString = tblBldr.String()
	} else {
		var (
			tblBldr = &strings.Builder{}
			tbl     = tablewriter.NewWriter(tblBldr)
		)
		tbl.SetHeader(t.headers)
		t.ApplyOpts(tbl)
		tblRows := []string{}
		for _, it := range t.innerTables {
			tblRows = append(tblRows, it.Render())
		}

		tbl.AppendBulk([][]string{tblRows})
		tbl.Render()
		returnString = tblBldr.String()
	}

	return returnString
}

func (t *Table) Headers() []string {
	var newHeaders []string
	for _, h := range t.headers {
		newHeaders = append(newHeaders, strings.Replace(h, "\n", "", -1)+"\n")
	}
	return newHeaders
}

func (t *Table) Data() [][]string {
	var newData [][]string
	for _, d := range t.data {
		var newInnerData []string
		for _, id := range d {
			newInnerData = append(newInnerData, strings.Replace(id, "\n", "", -1))
		}
		newData = append(newData, newInnerData)
	}
	return newData
}

// Testing
func (t *Table) RenderAsCSV() string {
	outstring := ""
	if len(t.innerTables) == 0 {
		outstring = renderTableAsCSV(t.Headers(), t.Data())
	} else {
		csvOutRaw := []string{}

		for idx, it := range t.innerTables {
			csvOutRaw = append(csvOutRaw, t.Headers()[idx])
			csvOutRaw = append(csvOutRaw, it.RenderAsCSV())
			csvOutRaw = append(csvOutRaw, "\n")
		}
		outstring = strings.Join(csvOutRaw, "")
	}

	return outstring
}

// renderSimpleTable is used to render any simple table within the Lacework CLI,
// every command should leverage this function unless there are extra customizations
// required, if so, use instead Table. The benefit of this function
// is the ability to switch/update the look and feel of the human-readable format
// across the entire project
func renderSimpleTable(headers []string, data [][]string) string {
	var (
		table = Table{
			headers: headers,
			data: data,
		}
	)
	return table.Render()
}

type tableOption interface {
	apply(t *tablewriter.Table)
}

type tableFunc func(t *tablewriter.Table)

func (fn tableFunc) apply(t *tablewriter.Table) {
	fn(t)
}
