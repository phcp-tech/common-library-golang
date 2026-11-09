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

// Package file checks file uploads against an upload policy: the allowed
// content types, the size limit and the file name. It has no storage and no
// database: where the file goes, who may upload it, the object key and the
// upload record stay in the application.
package file

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrInvalid is wrapped by every error of Policy.Check.
var ErrInvalid = errors.New("file: invalid upload")

// MaxFilenameLength is the longest file name ValidFilename accepts, in characters.
const MaxFilenameLength = 255

// ImageTypes returns the image types safe to show inline: png, jpeg, webp and
// gif. SVG is left out because it can carry scripts.
func ImageTypes() map[string]string {
	return map[string]string{
		"image/png":  "png",
		"image/jpeg": "jpg",
		"image/webp": "webp",
		"image/gif":  "gif",
	}
}

// DocumentTypes returns the document types: pdf and the Office Open XML
// formats without macros (docx, xlsx, pptx). The legacy binary formats (.doc,
// .xls, .ppt) and the macro-enabled ones (.docm, ...) are left out.
func DocumentTypes() map[string]string {
	return map[string]string{
		"application/pdf": "pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   "docx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "xlsx",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": "pptx",
	}
}

// Policy is what one kind of upload accepts.
type Policy struct {
	Types       map[string]string // content type -> file extension, without the dot
	MaxSize     int64             // largest file in bytes
	Disposition string            // Content-Disposition of the stored object, e.g. "attachment"; empty shows it inline
}

// Check validates one upload and returns the file extension of contentType.
// size must be 1 to MaxSize bytes. filename is optional and must pass
// ValidFilename. Every error wraps ErrInvalid.
func (p Policy) Check(contentType string, size int64, filename string) (string, error) {
	ext, ok := p.Types[contentType]
	if !ok {
		return "", fmt.Errorf("%w: contentType must be one of %s", ErrInvalid, strings.Join(slices.Sorted(maps.Keys(p.Types)), ", "))
	}
	if size <= 0 || size > p.MaxSize {
		return "", fmt.Errorf("%w: size must be 1 to %d bytes", ErrInvalid, p.MaxSize)
	}
	if !ValidFilename(filename) {
		return "", fmt.Errorf("%w: filename must be at most %d characters without control characters or path separators", ErrInvalid, MaxFilenameLength)
	}
	return ext, nil
}

// ValidFilename accepts an empty name, or valid UTF-8 of at most
// MaxFilenameLength characters without control characters, '/' or '\'.
// The name is for display and download only; never build a path or an
// object key from it.
func ValidFilename(name string) bool {
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) > MaxFilenameLength {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == '/' || r == '\\' {
			return false
		}
	}
	return true
}
