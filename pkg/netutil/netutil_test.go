// Copyright 2015 CoreOS, Inc.
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

package netutil

import (
	"errors"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestResolveTCPAddrs(t *testing.T) {
	defer func() { resolveTCPAddr = net.ResolveTCPAddr }()
	tests := []struct {
		urls     [][]url.URL
		expected [][]url.URL
		hostMap  map[string]string
		hasError bool
	}{
		{
			urls: [][]url.URL{
				[]url.URL{
					url.URL{Scheme: "http", Host: "127.0.0.1:4001"},
					url.URL{Scheme: "http", Host: "127.0.0.1:2379"},
				},
				[]url.URL{
					url.URL{Scheme: "http", Host: "127.0.0.1:7001"},
					url.URL{Scheme: "http", Host: "127.0.0.1:2380"},
				},
			},
			expected: [][]url.URL{
				[]url.URL{
					url.URL{Scheme: "http", Host: "127.0.0.1:4001"},
					url.URL{Scheme: "http", Host: "127.0.0.1:2379"},
				},
				[]url.URL{
					url.URL{Scheme: "http", Host: "127.0.0.1:7001"},
					url.URL{Scheme: "http", Host: "127.0.0.1:2380"},
				},
			},
		},
		{
			urls: [][]url.URL{
				[]url.URL{
					url.URL{Scheme: "http", Host: "infra0.example.com:4001"},
					url.URL{Scheme: "http", Host: "infra0.example.com:2379"},
				},
				[]url.URL{
					url.URL{Scheme: "http", Host: "infra0.example.com:7001"},
					url.URL{Scheme: "http", Host: "infra0.example.com:2380"},
				},
			},
			expected: [][]url.URL{
				[]url.URL{
					url.URL{Scheme: "http", Host: "10.0.1.10:4001"},
					url.URL{Scheme: "http", Host: "10.0.1.10:2379"},
				},
				[]url.URL{
					url.URL{Scheme: "http", Host: "10.0.1.10:7001"},
					url.URL{Scheme: "http", Host: "10.0.1.10:2380"},
				},
			},
			hostMap: map[string]string{
				"infra0.example.com": "10.0.1.10",
			},
			hasError: false,
		},
		{
			urls: [][]url.URL{
				[]url.URL{
					url.URL{Scheme: "http", Host: "infra0.example.com:4001"},
					url.URL{Scheme: "http", Host: "infra0.example.com:2379"},
				},
				[]url.URL{
					url.URL{Scheme: "http", Host: "infra0.example.com:7001"},
					url.URL{Scheme: "http", Host: "infra0.example.com:2380"},
				},
			},
			hostMap: map[string]string{
				"infra0.example.com": "",
			},
			hasError: true,
		},
		{
			urls: [][]url.URL{
				[]url.URL{
					url.URL{Scheme: "http", Host: "ssh://infra0.example.com:4001"},
					url.URL{Scheme: "http", Host: "ssh://infra0.example.com:2379"},
				},
				[]url.URL{
					url.URL{Scheme: "http", Host: "ssh://infra0.example.com:7001"},
					url.URL{Scheme: "http", Host: "ssh://infra0.example.com:2380"},
				},
			},
			hasError: true,
		},
	}
	for _, tt := range tests {
		resolveTCPAddr = func(network, addr string) (*net.TCPAddr, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			if tt.hostMap[host] == "" {
				return nil, errors.New("cannot resolve host.")
			}
			i, err := strconv.Atoi(port)
			if err != nil {
				return nil, err
			}
			return &net.TCPAddr{IP: net.ParseIP(tt.hostMap[host]), Port: i, Zone: ""}, nil
		}
		err := ResolveTCPAddrs(tt.urls...)
		if tt.hasError {
			if err == nil {
				t.Errorf("expected error")
			}
			continue
		}
		if !reflect.DeepEqual(tt.urls, tt.expected) {
			t.Errorf("expected: %v, got %v", tt.expected, tt.urls)
		}
	}
}

func TestURLsEqual(t *testing.T) {
	defer func() { resolveTCPAddr = net.ResolveTCPAddr }()
	resolveTCPAddr = func(network, addr string) (*net.TCPAddr, error) {
		host, port, err := net.SplitHostPort(addr)
		if host != "example.com" {
			return nil, errors.New("cannot resolve host.")
		}
		i, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		return &net.TCPAddr{IP: net.ParseIP("10.0.10.1"), Port: i, Zone: ""}, nil
	}

	tests := []struct {
		a      []url.URL
		b      []url.URL
		expect bool
	}{
		{
			a:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}},
			b:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}},
			expect: true,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "example.com:4001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.10.1:4001"}},
			expect: true,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: true,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "example.com:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "example.com:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: true,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "10.0.10.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "example.com:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: true,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}},
			b:      []url.URL{{Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "example.com:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.10.1:4001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.0.1:4001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "example.com:4001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.0.1:4001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "127.0.0.1:7001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "example.com:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "127.0.0.1:7001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "127.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "example.com:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: false,
		},
		{
			a:      []url.URL{{Scheme: "http", Host: "10.0.0.1:4001"}},
			b:      []url.URL{{Scheme: "http", Host: "10.0.0.1:4001"}, {Scheme: "http", Host: "127.0.0.1:7001"}},
			expect: false,
		},
	}

	for _, test := range tests {
		result := URLsEqual(test.a, test.b)
		if result != test.expect {
			t.Errorf("a:%v b:%v, expected %v but %v", test.a, test.b, test.expect, result)
		}
	}
}
func TestURLStringsEqual(t *testing.T) {
	result := URLStringsEqual([]string{"http://127.0.0.1:8080"}, []string{"http://127.0.0.1:8080"})
	if !result {
		t.Errorf("unexpected result %v", result)
	}
}
