// All modifications to this file are licensed under the LICENSE found at
// the root of the repository.
//
// Some contents are licensed under Apache Version 2.0 per below:
//
// Copyright 2018 Sourcegraph, Inc.
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

package bonsai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func AssertGolden(t testing.TB, path string, update bool, want any) {
	t.Helper()

	data := marshal(t, want)

	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
		if err := os.WriteFile(path, data, 0o640); err != nil { //nolint:gosec // writing files for VCR reading
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
	}

	golden, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %s", path, err)
	}

	if diff := cmp.Diff(string(golden), string(data)); diff != "" {
		t.Errorf("(-want, +got):\n%s", diff)
	}
}

func marshal(t testing.TB, v any) []byte {
	t.Helper()

	switch v2 := v.(type) {
	case string:
		return []byte(v2)
	case []byte:
		return v2
	default:
		data, err := json.MarshalIndent(v, " ", " ")
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
}