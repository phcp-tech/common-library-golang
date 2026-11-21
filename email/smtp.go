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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
)

// Sender sends one email message.
type Sender interface {
	Send(msg *Message) error
}

// Config is the SMTP server and the default sender. email/component reads it
// from the [email] section of app.toml.
type Config struct {
	Host        string // SMTP server host name
	Port        int    // 465 = implicit TLS; any other port must offer STARTTLS
	Username    string // SMTP login
	Password    string // SMTP password
	FromAddress string // sender address used when Message.From is empty
	FromName    string // optional display name of FromAddress
}

// Validate checks that the server, the login and the sender are set and well formed.
func (c Config) Validate() error {
	switch {
	case strings.TrimSpace(c.Host) == "":
		return errors.New("email: host must be set")
	case c.Port < 1 || c.Port > 65535:
		return fmt.Errorf("email: port %d is not a valid port number", c.Port)
	case c.Username == "" || c.Password == "":
		return errors.New("email: username and password must be set")
	case !IsEmail(c.FromAddress):
		return fmt.Errorf("email: from address %q is not a valid email address", c.FromAddress)
	}
	return nil
}

// sendTimeout limits one Send: connect, TLS, auth and the data transfer.
const sendTimeout = 30 * time.Second

// SMTPSender sends email through one SMTP server with github.com/wneessen/go-mail.
// It is safe for concurrent use.
type SMTPSender struct {
	cfg  Config
	send func(ctx context.Context, m *mail.Msg) error // replaced in tests
}

// NewSMTPSender checks cfg and returns a sender for it. It does not connect:
// each Send opens its own connection.
func NewSMTPSender(cfg Config) (*SMTPSender, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	s := &SMTPSender{cfg: cfg}
	s.send = func(ctx context.Context, m *mail.Msg) error {
		// A new client per message: a mail.Client holds one connection and is
		// not meant to be shared by concurrent sends.
		c, err := mail.NewClient(cfg.Host, clientOptions(cfg)...)
		if err != nil {
			return err
		}
		return c.DialAndSendWithContext(ctx, m)
	}
	return s, nil
}

// clientOptions maps cfg to go-mail options: port 465 uses implicit TLS, any
// other port must offer STARTTLS (TLSMandatory, so the password is never sent
// unencrypted). The auth mechanism is chosen from what the server offers
// (SCRAM, CRAM-MD5, PLAIN, LOGIN; PLAIN and LOGIN only over TLS).
func clientOptions(cfg Config) []mail.Option {
	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		mail.WithTimeout(mail.DefaultTimeout),
	}
	if cfg.Port == mail.DefaultPortSSL {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}
	return opts
}

// Send checks msg, builds the MIME message and sends it. An empty msg.From
// means the sender of the Config.
func (s *SMTPSender) Send(msg *Message) error {
	m, err := s.build(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	if err := s.send(ctx, m); err != nil {
		slog.Error("Send email failed", "to", msg.To[0], "server", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port), "error", err)
		return fmt.Errorf("send email: %w", err)
	}
	slog.Info("Email sent successfully", "to", msg.To[0])
	return nil
}

// bodyType returns the MIME type for msg.ContentType: text/plain or text/html,
// text/html when empty. Parameters such as charset are dropped, because the
// message is always encoded as UTF-8.
func bodyType(contentType string) (mail.ContentType, error) {
	if strings.TrimSpace(contentType) == "" {
		return mail.TypeTextHTML, nil
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("invalid contentType %q: %w", contentType, err)
	}
	switch mediaType {
	case "text/plain":
		return mail.TypeTextPlain, nil
	case "text/html":
		return mail.TypeTextHTML, nil
	}
	return "", fmt.Errorf("contentType %q is not supported, use text/plain or text/html", contentType)
}

// build checks msg and converts it to a go-mail message. go-mail also parses
// every address, so an invalid one is an error here, before any connection.
func (s *SMTPSender) build(msg *Message) (*mail.Msg, error) {
	if msg == nil {
		return nil, errors.New("email message is empty")
	}
	from, name := s.cfg.FromAddress, s.cfg.FromName
	if len(msg.From) > 0 {
		from, name = msg.From[0], ""
		if len(msg.From) > 1 {
			name = msg.From[1]
		}
	}
	if strings.TrimSpace(from) == "" {
		return nil, errors.New("email sender address must be set")
	}
	if len(msg.To) == 0 {
		return nil, errors.New("email receiver (to) must be set")
	}
	for _, list := range [][]string{msg.To, msg.Cc, msg.Bcc} {
		for _, addr := range list {
			if strings.TrimSpace(addr) == "" {
				return nil, errors.New("email address must not be empty")
			}
		}
	}
	contentType, err := bodyType(msg.ContentType)
	if err != nil {
		return nil, err
	}

	// UTF-8, quoted-printable, and no User-Agent/X-Mailer header naming the library.
	m := mail.NewMsg(mail.WithNoDefaultUserAgent())
	if strings.TrimSpace(name) != "" {
		err = m.FromFormat(name, from)
	} else {
		err = m.From(from)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}
	if err := m.To(msg.To...); err != nil {
		return nil, fmt.Errorf("invalid to address: %w", err)
	}
	if len(msg.Cc) > 0 {
		if err := m.Cc(msg.Cc...); err != nil {
			return nil, fmt.Errorf("invalid cc address: %w", err)
		}
	}
	if len(msg.Bcc) > 0 {
		if err := m.Bcc(msg.Bcc...); err != nil {
			return nil, fmt.Errorf("invalid bcc address: %w", err)
		}
	}
	m.Subject(msg.Subject)
	m.SetBodyString(contentType, msg.Body)
	return m, nil
}
