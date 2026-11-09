// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"errors"
	"strings"
	"testing"
)

func TestPolicy_Check(t *testing.T) {
	p := Policy{Types: ImageTypes(), MaxSize: 100}

	if ext, err := p.Check("image/jpeg", 100, strings.Repeat("名", MaxFilenameLength)); err != nil || ext != "jpg" {
		t.Fatalf("at the limits: ext %q, err %v", ext, err)
	}
	if ext, err := p.Check("image/png", 1, ""); err != nil || ext != "png" {
		t.Fatalf("no filename: ext %q, err %v", ext, err)
	}

	for name, c := range map[string]struct {
		ct       string
		size     int64
		filename string
	}{
		"svg":               {"image/svg+xml", 10, ""},
		"pdf in images":     {"application/pdf", 10, ""},
		"empty":             {"image/png", 0, ""},
		"too large":         {"image/png", 101, ""},
		"slash":             {"image/png", 10, "../a.png"},
		"backslash":         {"image/png", 10, `..\a.png`},
		"control character": {"image/png", 10, "a\nb.png"},
		"too long":          {"image/png", 10, strings.Repeat("名", MaxFilenameLength+1)},
		"invalid UTF-8":     {"image/png", 10, "a\xffb.png"},
	} {
		if _, err := p.Check(c.ct, c.size, c.filename); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

func TestTypes(t *testing.T) {
	if _, ok := ImageTypes()["image/svg+xml"]; ok {
		t.Error("ImageTypes allows SVG")
	}
	docs := DocumentTypes()
	if len(docs) != 4 || docs["application/pdf"] != "pdf" {
		t.Errorf("DocumentTypes = %v", docs)
	}
	if _, ok := docs["application/msword"]; ok {
		t.Error("DocumentTypes allows the legacy .doc format")
	}

	// each call returns a new map, so a caller cannot change another's policy
	ImageTypes()["image/svg+xml"] = "svg"
	if _, ok := ImageTypes()["image/svg+xml"]; ok {
		t.Error("ImageTypes returned a shared map")
	}
}
