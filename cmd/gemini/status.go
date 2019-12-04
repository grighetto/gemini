// Copyright 2019 ScyllaDB
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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/scylladb/gemini"
	"go.uber.org/zap"
)

type Status struct {
	WriteOps    int        `json:"write_ops"`
	WriteErrors int        `json:"write_errors"`
	ReadOps     int        `json:"read_ops"`
	ReadErrors  int        `json:"read_errors"`
	Errors      []JobError `json:"errors,omitempty"`
}

func (r *Status) Merge(sum *Status) Status {
	sum.WriteOps += r.WriteOps
	sum.WriteErrors += r.WriteErrors
	sum.ReadOps += r.ReadOps
	sum.ReadErrors += r.ReadErrors
	sum.Errors = append(sum.Errors, r.Errors...)
	return *sum
}

func (r *Status) PrintResult(w io.Writer, schema *gemini.Schema) {
	if err := r.PrintResultAsJSON(w, schema); err != nil {
		// In case there has been it has been a long run we want to display it anyway...
		fmt.Printf("Unable to print result as json, using plain text to stdout, error=%s\n", err)
		fmt.Printf("Gemini version: %s\n", version)
		fmt.Printf("Results:\n")
		fmt.Printf("\twrite ops:    %v\n", r.WriteOps)
		fmt.Printf("\tread ops:     %v\n", r.ReadOps)
		fmt.Printf("\twrite errors: %v\n", r.WriteErrors)
		fmt.Printf("\tread errors:  %v\n", r.ReadErrors)
		for i, err := range r.Errors {
			fmt.Printf("Error %d: %s\n", i, err)
		}
		jsonSchema, _ := json.MarshalIndent(schema, "", "    ")
		fmt.Printf("Schema: %v\n", string(jsonSchema))
	}
}

func (r *Status) PrintResultAsJSON(w io.Writer, schema *gemini.Schema) error {
	result := map[string]interface{}{
		"result":         r,
		"gemini_version": version,
		"schema":         schema,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", "    ")
	if err := encoder.Encode(result); err != nil {
		return errors.Wrap(err, "unable to create json from result")
	}
	return nil
}

func (r Status) String() string {
	return fmt.Sprintf("write ops: %v | read ops: %v | write errors: %v | read errors: %v", r.WriteOps, r.ReadOps, r.WriteErrors, r.ReadErrors)
}

func (r Status) HasErrors() bool {
	return r.WriteErrors > 0 || r.ReadErrors > 0
}

func sampleStatus(c chan Status, sp *spinner.Spinner, logger *zap.Logger) Status {
	failfastDone := sync.Once{}
	logger = logger.Named("sample_results")
	var testRes Status
	for res := range c {
		testRes = res.Merge(&testRes)
		if sp != nil {
			sp.Suffix = fmt.Sprintf(" Running Gemini... %v", testRes)
		}
		if testRes.ReadErrors > 0 || testRes.WriteErrors > 0 {
			if failFast {
				failfastDone.Do(func() {
					logger.Warn("Errors detected. Exiting.")
				})
				return testRes
			}
		}
	}
	return testRes
}
