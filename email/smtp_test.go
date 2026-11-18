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

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/wneessen/go-mail"
)

func validConfig() Config {
	return Config{Host: "smtp.example.test", Port: 587, Username: "user", Password: "secret",
		FromAddress: "noreply@example.test", FromName: "Toolmaker"}
}

// testSender returns a sender for validConfig whose send records the messages.
func testSender(t *testing.T) (*SMTPSender, *[]*mail.Msg) {
	t.Helper()
	s, err := NewSMTPSender(validConfig())
	if err != nil {
		t.Fatal(err)
	}
	var sent []*mail.Msg
	s.send = func(ctx context.Context, m *mail.Msg) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Error("send must get a context with a deadline")
		}
		sent = append(sent, m)
		return nil
	}
	return s, &sent
}

func validMessage() *Message {
	return &Message{
		From:    []string{"news@example.test", "News"},
		To:      []string{"a@example.test"},
		Subject: "Hello",
		Body:    "<p>Hi</p>",
	}
}

// render writes m as MIME text, so tests can check headers and body.
func render(t *testing.T, m *mail.Msg) string {
	t.Helper()
	var buf bytes.Buffer
	if _, err := m.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestConfig_Validate(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("valid config: %v", err)
	}
	for name, mutate := range map[string]func(*Config){
		"no host":     func(c *Config) { c.Host = " " },
		"port 0":      func(c *Config) { c.Port = 0 },
		"port 70000":  func(c *Config) { c.Port = 70000 },
		"no username": func(c *Config) { c.Username = "" },
		"no password": func(c *Config) { c.Password = "" },
		"bad from":    func(c *Config) { c.FromAddress = "noreply" },
		"no from":     func(c *Config) { c.FromAddress = "" },
	} {
		c := validConfig()
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Errorf("%s: want an error", name)
		}
		if _, err := NewSMTPSender(c); err == nil {
			t.Errorf("%s: NewSMTPSender must refuse the config", name)
		}
	}
	c := validConfig()
	c.FromName = ""
	if err := c.Validate(); err != nil {
		t.Errorf("from name is optional: %v", err)
	}
}

func TestBuild_HeadersAndDefaultHTML(t *testing.T) {
	s, _ := testSender(t)
	msg := validMessage()
	msg.Cc = []string{"c1@example.test", "c2@example.test"}
	msg.Bcc = []string{"b@example.test"}
	m, err := s.build(msg)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.GetCc(); len(got) != 2 {
		t.Errorf("Cc = %v, want both addresses", got)
	}
	if got := m.GetBcc(); len(got) != 1 {
		t.Errorf("Bcc = %v", got)
	}
	out := render(t, m)
	for _, want := range []string{`From: "News" <news@example.test>`, "To: <a@example.test>", "Subject: Hello", "Content-Type: text/html; charset=UTF-8"} {
		if !strings.Contains(out, want) {
			t.Errorf("message has no %q:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"User-Agent:", "X-Mailer:", "Bcc:"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("message must not have %s:\n%s", unwanted, out)
		}
	}
}

func TestBuild_From(t *testing.T) {
	s, _ := testSender(t)
	for _, c := range []struct {
		from []string
		want string
	}{
		{nil, `From: "Toolmaker" <noreply@example.test>`}, // the sender of the Config
		{[]string{"news@example.test"}, "From: <news@example.test>"},
		{[]string{"news@example.test", " "}, "From: <news@example.test>"},
	} {
		msg := validMessage()
		msg.From = c.from
		m, err := s.build(msg)
		if err != nil {
			t.Fatalf("from %v: %v", c.from, err)
		}
		if out := render(t, m); !strings.Contains(out, c.want) {
			t.Errorf("from %v: message has no %q:\n%s", c.from, c.want, out)
		}
	}
}

func TestBuild_ContentType(t *testing.T) {
	s, _ := testSender(t)
	for _, c := range []struct{ in, want string }{
		{"text/plain", "Content-Type: text/plain; charset=UTF-8"},
		{"text/html;charset=utf-8", "Content-Type: text/html; charset=UTF-8"},
		{"TEXT/PLAIN; charset=iso-8859-1", "Content-Type: text/plain; charset=UTF-8"},
	} {
		msg := validMessage()
		msg.ContentType = c.in
		m, err := s.build(msg)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if out := render(t, m); !strings.Contains(out, c.want) {
			t.Errorf("%q: message has no %q", c.in, c.want)
		}
	}
	for _, bad := range []string{"application/json", "text/"} {
		msg := validMessage()
		msg.ContentType = bad
		if _, err := s.build(msg); err == nil {
			t.Errorf("%q: want an error", bad)
		}
	}
}

func TestBuild_InvalidMessage(t *testing.T) {
	s, _ := testSender(t)
	for name, mutate := range map[string]func(*Message){
		"empty from":  func(m *Message) { m.From = []string{" "} },
		"bad from":    func(m *Message) { m.From = []string{"not an address"} },
		"no to":       func(m *Message) { m.To = nil },
		"empty to":    func(m *Message) { m.To = []string{""} },
		"bad to":      func(m *Message) { m.To = []string{"a@example.test", "bad@@x"} },
		"empty cc":    func(m *Message) { m.Cc = []string{""} },
		"bad cc":      func(m *Message) { m.Cc = []string{"bad"} },
		"empty bcc":   func(m *Message) { m.Bcc = []string{"x@example.test", " "} },
		"bad content": func(m *Message) { m.ContentType = "image/png" },
	} {
		msg := validMessage()
		mutate(msg)
		if _, err := s.build(msg); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	if _, err := s.build(nil); err == nil {
		t.Error("nil message: want an error")
	}
}

func TestSend(t *testing.T) {
	s, sent := testSender(t)
	if err := s.Send(validMessage()); err != nil || len(*sent) != 1 {
		t.Fatalf("err = %v, sent = %d", err, len(*sent))
	}
	if err := s.Send(&Message{}); err == nil || len(*sent) != 1 {
		t.Errorf("invalid message: err = %v, sent = %d; want an error and nothing sent", err, len(*sent))
	}

	boom := errors.New("connection refused")
	s.send = func(context.Context, *mail.Msg) error { return boom }
	if err := s.Send(validMessage()); !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestClientOptions_Port(t *testing.T) {
	for _, port := range []int{465, 587, 25} {
		c := validConfig()
		c.Port = port
		client, err := mail.NewClient(c.Host, clientOptions(c)...)
		if err != nil {
			t.Fatalf("port %d: %v", port, err)
		}
		if got := client.ServerAddr(); got != "smtp.example.test:"+strconv.Itoa(port) {
			t.Errorf("port %d: server address = %q", port, got)
		}
	}
}

func TestDefault(t *testing.T) {
	if err := InitDefault(Config{}); err == nil {
		t.Error("invalid config: want an error")
	}
	if err := InitDefault(validConfig()); err != nil {
		t.Fatal(err)
	}
	if Default() == nil {
		t.Error("Default() = nil after InitDefault")
	}
}
