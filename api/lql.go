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

package api

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	reLQL                  string = `(?ms)^(\w+)\([^)]+\)\s*{`
	LQLQueryTranslateError string = "unable to translate query blob"
)

type LQLQuery struct {
	ID             string `json:"LQL_ID,omitempty"`
	StartTimeRange string `json:"START_TIME_RANGE,omitempty"`
	EndTimeRange   string `json:"END_TIME_RANGE,omitempty"`
	QueryText      string `json:"QUERY_TEXT"`
	// QueryBlob is a special string that supports type conversion
	// back and forth from LQL to JSON
	QueryBlob string `json:"-"`
}

func (q *LQLQuery) Translate() (returnError error) {
	// if QueryText is populated; return
	if q.QueryText != "" {
		return
	}

	// if QueryBlob is JSON
	var t LQLQuery

	if err := json.Unmarshal([]byte(q.QueryBlob), &t); err == nil {
		if q.StartTimeRange == "" {
			q.StartTimeRange = t.StartTimeRange
		}
		if q.EndTimeRange == "" {
			q.EndTimeRange = t.EndTimeRange
		}
		q.QueryText = t.QueryText
		return
	}

	// if QueryBlob is LQL
	if matched, _ := regexp.MatchString(reLQL, q.QueryBlob); matched {
		q.QueryText = q.QueryBlob
		return
	}

	return errors.New(LQLQueryTranslateError)
}

func (q *LQLQuery) Validate(allowEmptyTimes bool) error {
	// translate
	if err := q.Translate(); err != nil {
		return err
	}
	// validate range
	if err := q.ValidateRange(allowEmptyTimes); err != nil {
		return err
	}
	// validate query
	if q.QueryText == "" {
		return errors.New("query should not be empty")
	}
	return nil
}

func (q LQLQuery) ValidateRange(allowEmptyTimes bool) error {
	// validate start
	start, err := q.ParseTime(q.StartTimeRange)
	if err != nil {
		if q.StartTimeRange == "" && allowEmptyTimes {
			start = time.Unix(0, 0)
		} else if q.StartTimeRange == "" {
			return errors.New("start time must not be empty")
		} else {
			return err
		}
	}
	// validate end
	end, err := q.ParseTime(q.EndTimeRange)
	if err != nil {
		if q.EndTimeRange == "" && allowEmptyTimes {
			end = time.Now()
		} else if q.EndTimeRange == "" {
			return errors.New("end time must not be empty")
		} else {
			return err
		}
	}
	// validate range
	if start.After(end) {
		return errors.New("date range should have a start time before the end time")
	}
	return nil
}

func (q LQLQuery) ParseTime(t string) (time.Time, error) {
	// parse time as RFC3339
	if tim, err := time.Parse(time.RFC3339, t); err == nil {
		return tim, err
	}
	// parse time as millis
	if msInt, err := strconv.ParseInt(t, 10, 64); err == nil {
		return time.Unix(0, msInt*int64(time.Millisecond)), err
	}
	return time.Time{}, errors.New("unable to parse time (" + t + ")")
}

type LQLQueryResponse struct {
	Data    []LQLQuery `json:"data"`
	Ok      bool       `json:"ok"`
	Message string     `json:"message"`
}

// LQLService is a service that interacts with the LQL
// endpoints from the Lacework Server
type LQLService struct {
	client *Client
}

func (svc *LQLService) CreateQuery(query string) (
	response LQLQueryResponse,
	err error,
) {
	lqlQuery := LQLQuery{QueryBlob: query}
	if err = lqlQuery.Validate(true); err != nil {
		return
	}

	err = svc.client.RequestEncoderDecoder("POST", ApiLQL, lqlQuery, &response)
	return
}

func (svc *LQLService) GetQueries() (
	response LQLQueryResponse,
	err error,
) {
	return svc.GetQueryByID("")
}

func (svc *LQLService) GetQueryByID(queryID string) (
	response LQLQueryResponse,
	err error,
) {
	uri := ApiLQL

	if queryID != "" {
		uri += "?LQL_ID=" + url.QueryEscape(queryID)
	}

	err = svc.client.RequestDecoder("GET", uri, nil, &response)
	return
}

func (svc *LQLService) RunQuery(query, start, end string) (
	response map[string]interface{},
	err error,
) {
	lqlQuery := LQLQuery{
		StartTimeRange: start,
		EndTimeRange:   end,
		QueryBlob:      query,
	}
	if err = lqlQuery.Validate(false); err != nil {
		return
	}

	err = svc.client.RequestEncoderDecoder("POST", ApiLQLQuery, lqlQuery, &response)
	return
}
