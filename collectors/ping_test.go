// Copyright 2020 Trey Dockendorf
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collectors

import (
	"os/exec"
	"testing"
)

func TestParseMmhealth(t *testing.T) {
	execCommand = fakeExecCommand
	mockedStdout = `
prometheus               time=55.63 ms


---- ping statistics ----
1 replies max: 55.63 min: 55.63 avg: 55.63 
`
	defer func() { execCommand = exec.Command }()
	metric, err := ping("prometheus")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if val := metric.Status; val != 1 {
		t.Errorf("Unexpected Status got %v", val)
	}
	if val := metric.Time; val != 0.055630000000000006 {
		t.Errorf("Unexpected Time got %v", val)
	}
}
