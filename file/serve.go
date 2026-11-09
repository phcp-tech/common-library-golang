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
	"fmt"
	"mime"
	"net/http"
	"os"
)

// ErrNotFound is returned by ServeAttachment when name is missing, is a
// directory, or is outside root.
var ErrNotFound = errors.New("file: not found")

// ServeAttachment sends the file name, relative to the directory root, as a
// download (Content-Disposition: attachment, with the file name encoded for
// non-ASCII characters). It supports Range and If-Modified-Since requests.
// A Content-Type the caller set on w is kept; otherwise it is detected.
//
// name cannot leave root: "..", absolute paths and symbolic links that point
// outside are refused (os.Root). On an error nothing has been written, so the
// caller writes the error response in its own format: ErrNotFound for a
// missing file, a directory or a name outside root; any other error means
// root itself cannot be opened.
//
// In Gin: file.ServeAttachment(c.Writer, c.Request, dir, c.Query("file")).
func ServeAttachment(w http.ResponseWriter, r *http.Request, root, name string) error {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("file: open root %s: %w", root, err)
	}
	defer dir.Close()

	f, err := dir.Open(name)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": info.Name()})
	if disposition == "" {
		disposition = "attachment"
	}
	w.Header().Set("Content-Disposition", disposition)
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	return nil
}
