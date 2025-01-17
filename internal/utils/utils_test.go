/*
Copyright 2022 Nethermind

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestSkipLines(t *testing.T) {
	inputs := []struct {
		content string
		symbol  string
		result  string
	}{
		{},
		{"bbbb", "a", "bbbb"},
		{"aaaa", "a", ""},
		{"bbbb\naaaa", "a", "bbbb"},
		{"aaaa\nbbbb", "a", "bbbb"},
	}

	for _, input := range inputs {
		descr := fmt.Sprintf("SkipLines(%s, %s)", input.content, input.symbol)
		got := SkipLines(input.content, input.symbol)

		if got != input.result {
			t.Errorf("%s expected %s but got %s", descr, input.result, got)
		}
	}
}

func TestContains(t *testing.T) {
	inputs := [...]struct {
		list []string
		str  string
		want bool
	}{
		{},
		{[]string{"a", "A", "b", "B"}, "b", true},
		{[]string{"a", "A", "b", "B"}, "c", false},
		{[]string{}, "a", false},
	}

	for _, input := range inputs {
		descr := fmt.Sprintf("Contains(%s)", input.list)
		if got := Contains(input.list, input.str); got != input.want {
			t.Errorf("%s expected %t but got %t", descr, input.want, got)
		}
	}
}

func TestContainsOnly(t *testing.T) {
	inputs := [...]struct {
		list   []string
		target []string
		want   bool
	}{
		{[]string{}, []string{}, true},
		{[]string{}, []string{"a"}, true},
		{[]string{"b"}, []string{"a", "A", "b", "B"}, true},
		{[]string{"c"}, []string{"a", "A", "b", "B"}, false},
		{[]string{"a"}, []string{}, false},
		{[]string{"execution", "validator"}, []string{"execution", "consensus"}, false},
		{[]string{"execution", "consensus", "validator"}, []string{"execution", "consensus"}, false},
		{[]string{"execution", "validator"}, []string{"execution", "consensus", "validator"}, true},
		{[]string{"execution", "consensus", "validator"}, []string{"execution", "consensus", "validator"}, true},
	}

	for _, input := range inputs {
		descr := fmt.Sprintf("Contains(%s)", input.list)
		if got := ContainsOnly(input.list, input.target); got != input.want {
			t.Errorf("%s expected %t but got %t", descr, input.want, got)
		}
	}
}

func TestIsAddress(t *testing.T) {
	tcs := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"2131", false},
		{"dasd31gsd1231", false},
		{"0x2312313aaef2312312", false},
		{"0x5c00ABEf07604C59Ac72E859E5F93D5abZXCVF83", false},
		{"5c00ABEf07604C59Ac72E859E5F93D5ab8546F83", false},
		{"0x5c00ABEf07604C59Ac72E859E5F93D5ab8546F83", true},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("IsAddress(%s)", tc.input), func(t *testing.T) {
			if got := IsAddress(tc.input); got != tc.want {
				t.Errorf("got != want. Expected %v, got %v", tc.want, tc.input)
			}
		})
	}
}

func TestPortAvailable(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		}))
	defer server.Close()
	split := strings.Split(server.URL, ":")
	host, strPort := split[1][2:], split[2]
	port64, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		t.Fatalf("cannot convert http server port: %v", err)
	}
	port := uint16(port64)

	tcs := []struct {
		name string
		host string
		port uint16
		want bool
	}{
		{
			"Test case 1, good host and unavailable port",
			host, port,
			false,
		},
		{
			"Test case 2, bad host and port",
			"b@dh0$t", port,
			true,
		},
		{
			"Test case 3, good host and bad port",
			host, 9999,
			true,
		},
		{
			"Test case 4, good host and available port",
			"localhost", 9999,
			true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if got := portAvailable(tc.host, tc.port); tc.want != got {
				t.Errorf("portAvailable(%s, %d) failed; expected: %v, got: %v", tc.host, tc.port, tc.want, got)
			}
		})
	}
}

func TestAssignPorts(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		}))
	defer server.Close()
	split := strings.Split(server.URL, ":")
	host, strPort := split[1][2:], split[2]
	port64, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		t.Fatalf("cannot convert http server port: %v", err)
	}
	port := uint16(port64)

	tcs := []struct {
		name     string
		host     string
		defaults map[string]uint16
		want     map[string]uint16
		isErr    bool
	}{
		{
			"Test case 1, good host and defaults",
			host,
			map[string]uint16{"EL": 8545, "CL": port},
			map[string]uint16{"EL": 8545, "CL": port + 1},
			false,
		},
		{
			"Test case 2, good host and bad defaults",
			host,
			map[string]uint16{"EL": 8545, "CL": 0},
			map[string]uint16{},
			true,
		},
		{
			"Test case 3, good host and bad defaults",
			host,
			map[string]uint16{"CL": 0, "EL": 8545},
			map[string]uint16{},
			true,
		},
		{
			"Test case 4, bad host and good defaults",
			"b@dh0$t",
			map[string]uint16{"CL": 9000, "EL": 8545},
			map[string]uint16{"CL": 9000, "EL": 8545},
			false,
		},
		{
			"Test case 5, good host and successive increments",
			host,
			map[string]uint16{"CL": port, "EL": port + 1},
			map[string]uint16{"CL": port + 1, "EL": port + 2},
			false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AssignPorts(tc.host, tc.defaults)

			descr := fmt.Sprintf("AssingPorts(%s, %+v)", tc.host, tc.defaults)
			if cerr := CheckErr(descr, tc.isErr, err); cerr != nil {
				t.Error(cerr)
			}

			if err == nil {
				for k := range tc.want {
					if tc.want[k] != got[k] {
						t.Errorf("A mismatch in the result has been found. Expected (key: %s, value: %d); got (key: %s, value %d). Call: %s. Expected object: %+v, Got: %+v", k, tc.want[k], k, got[k], descr, tc.want, got)
					}
				}
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tcs := []struct {
		name   string
		in     []string
		want   []string
		filter func(string) bool
	}{
		{
			"Test case 1, no filter",
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			func(s string) bool {
				return true
			},
		},
		{
			"Test case 2, filter",
			[]string{"a", "b", "c"},
			[]string{"a", "c"},
			func(s string) bool {
				return s != "b"
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.in, tc.filter)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Filter(%+v) failed; expected: %+v, got: %+v", tc.in, tc.want, got)
			}
		})
	}
}
