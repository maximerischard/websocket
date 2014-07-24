// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"net/http"
	"reflect"
	"testing"
)

var subprotocolTests = []struct {
	h         string
	protocols []string
}{
	{"", nil},
	{"foo", []string{"foo"}},
	{"foo,bar", []string{"foo", "bar"}},
	{"foo, bar", []string{"foo", "bar"}},
	{" foo, bar", []string{"foo", "bar"}},
	{" foo, bar ", []string{"foo", "bar"}},
}

func TestSubprotocols(t *testing.T) {
	for _, st := range subprotocolTests {
		r := http.Request{Header: http.Header{"Sec-Websocket-Protocol": {st.h}}}
		protocols := Subprotocols(&r)
		if !reflect.DeepEqual(st.protocols, protocols) {
			t.Errorf("SubProtocols(%q) returned %#v, want %#v", st.h, protocols, st.protocols)
		}
	}
}

var extensionTests = []struct {
	h          string
	extensions ExtensionList
}{
	{"", nil},
	{"foo", ExtensionList{Extension{"foo", map[string]string{}}}},
	{"foo,bar;baz", ExtensionList{Extension{"foo", map[string]string{}}, Extension{"bar", map[string]string{"baz": ""}}}},
	{"foo,bar;baz=3", ExtensionList{Extension{"foo", map[string]string{}}, Extension{"bar", map[string]string{"baz": "3"}}}},
	{"foo, bar", ExtensionList{Extension{"foo", map[string]string{}}, Extension{"bar", map[string]string{}}}},
	{" foo, bar", ExtensionList{Extension{"foo", map[string]string{}}, Extension{"bar", map[string]string{}}}},
	{" foo, bar ", ExtensionList{Extension{"foo", map[string]string{}}, Extension{"bar", map[string]string{}}}},
	// and a real world example
	{"permessage-deflate; client_max_window_bits, x-webkit-deflate-frame", ExtensionList{Extension{"permessage-deflate", map[string]string{"client_max_window_bits": ""}}, Extension{"x-webkit-deflate-frame", map[string]string{}}}},
}

func TestExtensions(t *testing.T) {
	for _, extest := range extensionTests {
		r := http.Request{Header: http.Header{"Sec-Websocket-Extensions": {extest.h}}}
		extensions := Extensions(r.Header)
		if !reflect.DeepEqual(extest.extensions, extensions) {
			t.Errorf("Extensions(%q) returned %#v, want %#v", extest.h, extensions, extest.extensions)
		}
	}
}
