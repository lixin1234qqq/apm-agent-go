// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package stacktrace_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.elastic.co/apm/stacktrace"
)

func TestStacktrace(t *testing.T) {
	expect := []string{
		"go.elastic.co/apm/stacktrace_test.TestStacktrace.func1",
		"runtime.call32",
		"runtime.gopanic",
		"go.elastic.co/apm/stacktrace_test.(*panicker).panic",
		"go.elastic.co/apm/stacktrace_test.TestStacktrace",
	}
	defer func() {
		err := recover()
		if err == nil {
			t.FailNow()
		}
		allFrames := stacktrace.AppendStacktrace(nil, 1, 5)
		functions := make([]string, len(allFrames))
		for i, frame := range allFrames {
			functions[i] = frame.Function
		}
		if diff := cmp.Diff(functions, expect); diff != "" {
			t.Fatalf("%s", diff)
		}
	}()
	(&panicker{}).panic()
}

type panicker struct{}

func (*panicker) panic() {
	panic("oh noes")
}

func TestSplitFunctionName(t *testing.T) {
	testSplitFunctionName(t, "main", "main")
	testSplitFunctionName(t, "main", "Foo.Bar")
	testSplitFunctionName(t, "main", "(*Foo).Bar")
	testSplitFunctionName(t, "go.elastic.co/apm/foo", "bar")
	testSplitFunctionName(t,
		"go.elastic.co/apm/module/apmgin",
		"(*middleware).(go.elastic.co/apm/module/apmgin.handle)-fm",
	)
}

func testSplitFunctionName(t *testing.T, module, function string) {
	outModule, outFunction := stacktrace.SplitFunctionName(module + "." + function)
	assertModule(t, outModule, module)
	assertFunction(t, outFunction, function)
}

func TestSplitFunctionNameUnescape(t *testing.T) {
	module, function := stacktrace.SplitFunctionName("github.com/elastic/apm-agent%2ego.funcName")
	assertModule(t, module, "github.com/elastic/apm-agent.go")
	assertFunction(t, function, "funcName")

	// malformed escape sequences are left alone
	module, function = stacktrace.SplitFunctionName("github.com/elastic/apm-agent%.funcName")
	assertModule(t, module, "github.com/elastic/apm-agent%")
	assertFunction(t, function, "funcName")
}

func assertModule(t *testing.T, got, expect string) {
	if got != expect {
		t.Errorf("got module %q, expected %q", got, expect)
	}
}

func assertFunction(t *testing.T, got, expect string) {
	if got != expect {
		t.Errorf("got function %q, expected %q", got, expect)
	}
}
