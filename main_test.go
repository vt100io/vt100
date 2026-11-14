package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func get(t *testing.T, path string) (*http.Response, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	newSite().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)
	return resp, string(body)
}

func TestHomeListsEveryTool(t *testing.T) {
	resp, body := get(t, "/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	for _, tool := range tools {
		if !strings.Contains(body, `href="/tools/`+tool.Name+`"`) {
			t.Errorf("front page has no link to %s", tool.Name)
		}
	}
}

func TestToolPages(t *testing.T) {
	for _, tool := range tools {
		resp, body := get(t, "/tools/"+tool.Name)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d", tool.Name, resp.StatusCode)
			continue
		}
		for _, want := range []string{"<title>" + tool.Name + " · vt100</title>", tool.Install, tool.Repo()} {
			if !strings.Contains(body, want) {
				t.Errorf("%s page lacks %q", tool.Name, want)
			}
		}
	}
}

func TestGuides(t *testing.T) {
	_, home := get(t, "/")
	for _, g := range guides {
		if !strings.Contains(home, `href="/guides/`+g.Slug+`"`) {
			t.Errorf("front page has no link to guide %s", g.Slug)
		}
		resp, body := get(t, "/guides/"+g.Slug)
		if resp.StatusCode != http.StatusOK || !strings.Contains(body, "<title>"+g.Title+" · vt100</title>") {
			t.Errorf("guide %s: status %d", g.Slug, resp.StatusCode)
		}
	}
}

func TestNotFound(t *testing.T) {
	for _, path := range []string{"/tools/nope", "/guides/nope", "/nope", "/static/nope.css"} {
		if resp, _ := get(t, path); resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, resp.StatusCode)
		}
	}
}

// Every file a page refers to is embedded, so the site works with nothing
// but the binary.
func TestPagesOnlyReferToEmbeddedFiles(t *testing.T) {
	_, css := get(t, "/static/site.css")
	pages := []string{"/", "/tools/punchcard"}
	var refs []string
	for _, p := range pages {
		_, body := get(t, p)
		refs = append(refs, attrValues(body, `src="`, "")...)
		refs = append(refs, attrValues(body, `href="/static/`, "/static/")...)
	}
	refs = append(refs, attrValues(css, `url("`, "")...)
	for _, ref := range refs {
		if !strings.HasPrefix(ref, "/") {
			ref = "/static/" + ref // relative to site.css
		}
		if resp, _ := get(t, ref); resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d", ref, resp.StatusCode)
		}
	}
	if len(refs) < 7 { // the demo gif, the stylesheet and six fonts at least
		t.Errorf("found only %d references: %v", len(refs), refs)
	}
}

func TestSecurityHeaders(t *testing.T) {
	resp, _ := get(t, "/")
	if csp := resp.Header.Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("CSP = %q", csp)
	}
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff")
	}
}

// attrValues returns what follows each occurrence of prefix up to the next
// quote, with keep put back in front.
func attrValues(s, prefix, keep string) []string {
	var out []string
	for {
		i := strings.Index(s, prefix)
		if i < 0 {
			return out
		}
		s = s[i+len(prefix):]
		end := strings.IndexByte(s, '"')
		if end < 0 {
			return out
		}
		out = append(out, keep+s[:end])
		s = s[end:]
	}
}

// Under systemd the site says READY=1 and then keeps pinging the watchdog,
// or unit's Type=notify service is killed after its start timeout.
func TestNotifiesSystemd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notify.sock")
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: path, Net: "unixgram"})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	t.Setenv("NOTIFY_SOCKET", path)
	t.Setenv("WATCHDOG_USEC", "20000") // ping every 10ms

	notifySystemd("READY=1")
	go watchdog()

	buf := make([]byte, 64)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var got []string
	for len(got) < 3 {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("after %v: %v", got, err)
		}
		got = append(got, string(buf[:n]))
	}
	if got[0] != "READY=1" || got[1] != "WATCHDOG=1" || got[2] != "WATCHDOG=1" {
		t.Errorf("messages = %v", got)
	}
}
