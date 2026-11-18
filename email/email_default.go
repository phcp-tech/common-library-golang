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

package email

import "sync"

var (
	instance *SMTPSender
	mu       sync.RWMutex
)

// InitDefault creates the package-level default sender from cfg.
// A later successful call replaces the previous default.
func InitDefault(cfg Config) error {
	s, err := NewSMTPSender(cfg)
	if err != nil {
		return err
	}
	mu.Lock()
	instance = s
	mu.Unlock()
	return nil
}

// Default returns the package-level default sender, or nil before InitDefault
// succeeds. Use NewSMTPSender for any other instance.
func Default() *SMTPSender {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}
