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

// Package download serves files under one directory as downloads on a Gin
// engine. The file checks live in the Gin-free file package
// (file.ServeAttachment); this package adds the Gin handler, the route and the
// JSON error responses.
package download

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phcp-tech/common-library-golang/file"
	"github.com/phcp-tech/common-library-golang/log"

	"github.com/gin-gonic/gin"
)

// ExeDir returns the directory of the running executable. Use it as the root
// for applications that run as a Windows service, whose working directory is
// not the executable's directory.
func ExeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// Handler serves the file named by the "file" query parameter, relative to
// root, as an attachment. The name must start with "./" and cannot leave root
// (see file.ServeAttachment); directories are not listed. A Content-Type the
// caller set before calling the handler is kept.
//
// Errors are JSON {"code", "message"}: 400 for a name without "./", 404 for a
// missing file, a directory or a name outside root, 500 when root cannot be
// opened. Content-Disposition is exposed to cross-origin scripts, so a browser
// client can read the file name.
func Handler(root string) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("file")
		if !strings.HasPrefix(name, "./") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Download file must from ./ folder",
			})
			return
		}

		log.Infof("Begin to download file: %s", name)
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")
		err := file.ServeAttachment(c.Writer, c.Request, root, name)
		switch {
		case errors.Is(err, file.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "File not found",
			})
		case err != nil:
			log.Errorf("Download file %s: %v", name, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Failed to open the download directory",
			})
		}
	}
}

// Mount registers GET <path>/admin/files/download, serving Handler(root)
// behind handlers, e.g. token.Authenticate(). It panics when handlers is
// empty: the route serves files from disk and must not be public by mistake.
func Mount(router *gin.Engine, path, root string, handlers ...gin.HandlerFunc) *gin.Engine {
	if len(handlers) == 0 {
		panic("download: Mount needs at least one handler, e.g. token.Authenticate()")
	}
	router.Group(path+"/admin/files", handlers...).GET("/download", Handler(root))
	return router
}
