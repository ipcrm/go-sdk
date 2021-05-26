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
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// Table struct used for both console tables as well as CSV.  Do NOT use directly, use renderSimpleTable,
// renderCustomTable, or NewTable
type Table struct {
	// Headers for each table field
	headers     []string

	// Data to be included in this table
	data        [][]string

	// If this table is wrapping other tables, supply those other tables here.
	// IMPORTANT: The headers for this table are auto-assigned to each inner table based on postion.
	// Example: 2 hears, 2 inner tables.  Inner table 1 is shown under header 1.
	innerTables []*Table

	// Supply table options to customize behavior.  See tablewriter interface for details.
	opts        []tableOption

	// Future use; you SHOULD populate this with a human readable, lowercase, hyphenated text, or set as "" to not be able
	// to use in csv printing
	label       string

	// Future use; should this table output be rendered as a CSV
	renderAsCsv bool
	csvSection  string
}

// Apply any custom options
func (t *Table) ApplyOpts(tbl *tablewriter.Table) {
	for _, opt := range t.opts {
		opt.apply(tbl)
	}
}

// Render table
func (t *Table) Render() string {
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

// Convert table to CSV format for output
func (t *Table) RenderAsCSV() string {

	// Build list of all labels
	validSections := []string{}

	if t.label != "" {
		validSections = append(validSections, t.label)
	}
	for _, it := range t.innerTables {
		if it.label != "" {
		  validSections = append(validSections, it.label)
		}
	}

	// Check if section has been set or if there is only one valid section just set it automatically
	if len(validSections) > 1 && t.csvSection == "" {
		fmt.Printf("You must supply a section when requesting CSV output. Use --csv_section. Valid sections are %v\n", strings.Join(validSections, ", "))
		os.Exit(1)
	} else if len(validSections) == 1 && t.csvSection == "" {
		t.csvSection = validSections[0]
	}

	// Find the table we are supposed to print
	// var data [][]string
	// var headers []string
	// if t.csvSection == t.label {
		// data = t.Data()
		// headers = t.Headers()
	// } else {
		// for _, it := range t.innerTables {
			// if it.label == t.csvSection {
				// data = it.Data()
				// headers = it.Headers()
			// }
		// }
	// }

	var outstring string
	if t.label == t.csvSection && len(t.innerTables) == 0 {
		outstring = renderTableAsCSV(t.Headers(), t.Data())
	} else {
		csvOutRaw := []string{}

		for idx, it := range t.innerTables {
			if it.label == t.csvSection {
				if len(t.Headers()) < idx {
					csvOutRaw = append(csvOutRaw, t.Headers()[idx])
				}
				csvOutRaw = append(csvOutRaw, it.RenderAsCSV())
				csvOutRaw = append(csvOutRaw, "\n")
			}
		}
		outstring = strings.Join(csvOutRaw, "")
	}

	// return renderTableAsCSV(headers, data)
	return outstring
}

type tableOption interface {
	apply(t *tablewriter.Table)
}

type tableFunc func(t *tablewriter.Table)

func (fn tableFunc) apply(t *tablewriter.Table) {
	fn(t)
}

func NewTable(label string, headers []string, data [][]string, innerTables []*Table, opts ...tableOption) *Table {
	// TODO This an OK way to do this validation?
	if innerTables != nil && data != nil {
		log.Fatal("Cannot supply both innerTables and data to NewTable")
	}

	t := new(Table)
	t.headers = headers
	t.innerTables = innerTables
	t.data = data
	t.opts = opts
	t.label = label


	// TODO Acceptable access?
	if cli.csvOutput {
		t.renderAsCsv = true
		t.csvSection = cli.csvSection
	}

	return t
}

// renderCustomTable is used to render tables that have more complex requirements than is possible with
// renderSimpleTable.  renderCustomTable allows supplying visual format customizations to the underlying tablewriter
// library via `opts`.  In addition, tables can be nested using renderCustomTable via `innerTables`.  Tables to be
// passed to `innerTables` should be created via `NewTable`
func renderCustomTable(label string, headers []string, data [][]string, innerTables []*Table, opts ...tableOption) string {
	return NewTable(label, headers, data, innerTables, opts...).Render()
}

// renderSimpleTable is used to render any simple table within the Lacework CLI,
// every command should leverage this function unless there are extra customizations
// required, if so, use instead renderCustomTable or NewTable. The benefit of this function
// is the ability to switch/update the look and feel of the human-readable format
// across the entire project
func renderSimpleTable(headers []string, data [][]string) string {
	return renderCustomTable("", headers, data, nil)
}
