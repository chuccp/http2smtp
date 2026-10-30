package smtp

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/chuccp/http2smtp/model"
)

// fakeSMTP speaks just enough ESMTP for go-mail: EHLO/HELO, AUTH PLAIN (with or without
// an initial response), MAIL, RCPT, DATA, RSET, QUIT. Recipients listed in reject get
// the configured SMTP response instead of an acceptance, which is how the send paths
// that react to a "user not found" answer are exercised without a real mail server.
type fakeSMTP struct {
	ln         net.Listener
	mu         sync.Mutex
	sessions   []*fakeSession
	reject     map[string]string // recipient address -> full SMTP error line
	rejectData string            // if set, the answer sent instead of accepting DATA
}

type fakeSession struct {
	from      string
	attempted []string // every RCPT TO the server saw
	accepted  []string
	data      string
}

func newFakeSMTP(t *testing.T, reject map[string]string, rejectData string) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	f := &fakeSMTP{ln: ln, reject: reject, rejectData: rejectData}
	go f.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return f
}

func (f *fakeSMTP) port() int {
	return f.ln.Addr().(*net.TCPAddr).Port
}

func (f *fakeSMTP) all() []*fakeSession {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*fakeSession(nil), f.sessions...)
}

func (f *fakeSMTP) serve() {
	for {
		conn, err := f.ln.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *fakeSMTP) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	reply := func(format string, args ...any) {
		_, _ = fmt.Fprintf(w, format, args...)
		_ = w.Flush()
	}
	reply("220 fake ESMTP\r\n")

	var session *fakeSession
	inData := false
	var data strings.Builder
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		if inData {
			if line == ".\r\n" {
				inData = false
				if session != nil {
					session.data = data.String()
				}
				if f.rejectData != "" {
					reply("%s\r\n", f.rejectData)
				} else {
					reply("250 2.0.0 ok\r\n")
				}
			} else {
				data.WriteString(line)
			}
			continue
		}
		upper := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(upper, "EHLO"):
			reply("250-fake\r\n250-AUTH PLAIN LOGIN\r\n250 SIZE 10000000\r\n")
		case strings.HasPrefix(upper, "HELO"):
			reply("250 fake\r\n")
		case strings.HasPrefix(upper, "AUTH PLAIN"):
			// Accept both the one-line initial response and the challenge form.
			if len(strings.TrimSpace(line)) == len("AUTH PLAIN") {
				reply("334 \r\n")
				if _, err := r.ReadString('\n'); err != nil {
					return
				}
			}
			reply("235 2.7.0 ok\r\n")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			session = &fakeSession{from: extractAddress(line)}
			f.mu.Lock()
			f.sessions = append(f.sessions, session)
			f.mu.Unlock()
			reply("250 2.1.0 ok\r\n")
		case strings.HasPrefix(upper, "RCPT TO:"):
			addr := extractAddress(line)
			if session != nil {
				session.attempted = append(session.attempted, addr)
			}
			if msg, bad := f.reject[addr]; bad {
				reply("%s\r\n", msg)
				continue
			}
			if session != nil {
				session.accepted = append(session.accepted, addr)
			}
			reply("250 2.1.5 ok\r\n")
		case upper == "DATA":
			inData = true
			data.Reset()
			reply("354 end with .\r\n")
		case upper == "RSET":
			reply("250 2.0.0 ok\r\n")
		case upper == "NOOP":
			reply("250 2.0.0 ok\r\n")
		case upper == "QUIT":
			reply("221 2.0.0 bye\r\n")
			return
		default:
			reply("502 5.5.2 not implemented\r\n")
		}
	}
}

// extractAddress pulls the addr-spec out of "RCPT TO:<a@example.com>" style lines.
func extractAddress(line string) string {
	if i := strings.Index(line, "<"); i >= 0 {
		if j := strings.LastIndex(line, ">"); j > i {
			return line[i+1 : j]
		}
	}
	// Bare "MAIL FROM:a@example.com" or trailing SIZE=...
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		addr := strings.TrimSpace(parts[1])
		if k := strings.Index(addr, " "); k >= 0 {
			addr = addr[:k]
		}
		return strings.Trim(addr, "<>")
	}
	return ""
}

func testModelSMTP(port int) *model.SMTP {
	return &model.SMTP{Host: "127.0.0.1", Port: port, Mail: "sender@example.com", Username: "sender", Password: "pw"}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// A plain send reaches every accepted recipient in one transaction, with the envelope
// sender the SMTP row carries and the subject/body/attachment in the message.
func TestSendFilesMsgDeliversToAllRecipients(t *testing.T) {
	f := newFakeSMTP(t, nil, "")
	handle, err := os.Open(writeTempFile(t, "attach.txt", "attachment-bytes"))
	if err != nil {
		t.Fatalf("open attachment: %v", err)
	}
	t.Cleanup(func() { _ = handle.Close() })
	file := &File{Name: "attach.txt", FilePath: handle.Name(), File: handle}

	mails := []*model.Mail{
		{Name: "A", Mail: "a@example.com"},
		{Name: "B", Mail: "b@example.com"},
	}
	if err := SendFilesMsg(testModelSMTP(f.port()), mails, []*File{file}, "wire subject", "wire body"); err != nil {
		t.Fatalf("send: %v", err)
	}
	sessions := f.all()
	if len(sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(sessions))
	}
	s := sessions[0]
	if s.from != "sender@example.com" {
		t.Errorf("envelope from = %q, want sender@example.com", s.from)
	}
	if strings.Join(s.attempted, ",") != "a@example.com,b@example.com" {
		t.Errorf("attempted recipients = %v", s.attempted)
	}
	if !strings.Contains(s.data, "Subject: wire subject") || !strings.Contains(s.data, "wire body") {
		t.Errorf("message does not carry the subject/body:\n%s", s.data)
	}
	// The attachment travels base64-encoded inside the MIME part.
	encoded := base64.StdEncoding.EncodeToString([]byte("attachment-bytes"))
	if !strings.Contains(s.data, `filename="attach.txt"`) || !strings.Contains(s.data, encoded) {
		t.Errorf("message does not carry the attachment, got:\n%s", s.data)
	}
}

// The "user not found" answer triggers one retry without the rejected addresses: the
// remaining recipients are still delivered, and the caller gets the user-not-found
// marker that maps the log entry to a warning.
func TestSendMailRetriesDroppingUnknownUser(t *testing.T) {
	f := newFakeSMTP(t, map[string]string{
		"b@example.com": "550 5.1.1 <b@example.com>: user not found",
	}, "")
	mails := []*model.Mail{
		{Name: "A", Mail: "a@example.com"},
		{Name: "B", Mail: "b@example.com"},
	}
	err := SendFilesMsg(testModelSMTP(f.port()), mails, nil, "sub", "body")
	if err == nil {
		t.Fatal("expected the send to report the unknown user")
	}
	var uerr *UserNotFoundError
	if !errors.As(err, &uerr) {
		t.Fatalf("error = %v (%T), want *UserNotFoundError", err, err)
	}
	if !strings.Contains(err.Error(), "b@example.com") {
		t.Errorf("error = %v, want it to name b@example.com", err)
	}
	sessions := f.all()
	if len(sessions) != 2 {
		t.Fatalf("sessions = %d, want the original plus one retry", len(sessions))
	}
	if strings.Join(sessions[0].attempted, ",") != "a@example.com,b@example.com" {
		t.Errorf("first attempt = %v, want both recipients", sessions[0].attempted)
	}
	if strings.Join(sessions[1].attempted, ",") != "a@example.com" {
		t.Errorf("retry = %v, want only the remaining recipient", sessions[1].attempted)
	}
	if !strings.Contains(sessions[1].data, "body") {
		t.Error("the retry did not deliver the message")
	}
}

// Any other failure is reported as-is: no retry, no user-not-found marker.
func TestSendMailDoesNotRetryOnOtherErrors(t *testing.T) {
	f := newFakeSMTP(t, map[string]string{
		"b@example.com": "550 5.7.1 relay denied",
	}, "")
	mails := []*model.Mail{
		{Name: "A", Mail: "a@example.com"},
		{Name: "B", Mail: "b@example.com"},
	}
	err := SendFilesMsg(testModelSMTP(f.port()), mails, nil, "sub", "body")
	if err == nil {
		t.Fatal("expected the send to fail")
	}
	var uerr *UserNotFoundError
	if errors.As(err, &uerr) {
		t.Errorf("error = %v, want a plain failure without the user-not-found marker", err)
	}
	if sessions := f.all(); len(sessions) != 1 {
		t.Errorf("sessions = %d, want 1 — a non-'not found' error must not retry", len(sessions))
	}
}

// A "not found" answer that names no address gives the retry nothing to exclude. The
// send must report the failure once — not resend the identical message, and certainly
// not recurse until the process dies. The retry branch re-enters itself whenever the
// excluded list is empty, so an empty extraction must never lead to a resend.
func TestSendMailNotFoundWithoutAddressDoesNotRetry(t *testing.T) {
	f := newFakeSMTP(t, nil, "550 5.7.1 policy: sender domain not found")
	mails := []*model.Mail{
		{Name: "A", Mail: "a@example.com"},
		{Name: "B", Mail: "b@example.com"},
	}
	err := SendFilesMsg(testModelSMTP(f.port()), mails, nil, "sub", "body")
	if err == nil {
		t.Fatal("expected the send to fail")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want the server answer", err)
	}
	var uerr *UserNotFoundError
	if errors.As(err, &uerr) {
		t.Errorf("error = %v, want a plain failure: no user could be identified", err)
	}
	if sessions := f.all(); len(sessions) != 1 {
		t.Errorf("sessions = %d, want 1 — an unattributable 'not found' must not retry", len(sessions))
	}
}
