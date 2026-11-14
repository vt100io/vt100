// vt100 serves the vt100 site: calm architecture, and the small tools I use
// to run my hobby projects. Pages, styles, fonts and images are embedded, so
// the binary is the whole site.
package main

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

func main() {
	l, err := listen()
	if err != nil {
		slog.Error("listening", "err", err)
		os.Exit(1)
	}
	srv := &http.Server{
		Handler:           newSite(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	slog.Info("vt100 listening", "addr", l.Addr())
	// unit runs services as Type=notify with a watchdog: report ready, then
	// keep reporting alive.
	notifySystemd("READY=1")
	go watchdog()
	if err := srv.Serve(l); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// page is one template set: the shared layout plus the page's own content.
func page(name string) *template.Template {
	funcs := template.FuncMap{"githubOwner": func() string { return githubOwner }}
	return template.Must(template.New(name).Funcs(funcs).ParseFS(templateFS, "templates/layout.html", "templates/"+name+".html"))
}

func newSite() http.Handler {
	home, tool, principles, notFound := page("home"), page("tool"), page("principles"), page("notfound")
	guidePages := map[string]*template.Template{}
	for _, g := range guides {
		guidePages[g.Slug] = page("guide-" + g.Slug)
	}
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", cacheFor(24*time.Hour, http.FileServerFS(static))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		render(w, home, http.StatusOK, map[string]any{"Tools": tools, "Guides": guides})
	})
	mux.HandleFunc("GET /principles", func(w http.ResponseWriter, r *http.Request) {
		render(w, principles, http.StatusOK, map[string]any{"Title": "Principles"})
	})
	mux.HandleFunc("GET /tools/{name}", func(w http.ResponseWriter, r *http.Request) {
		t, ok := toolByName(r.PathValue("name"))
		if !ok {
			render(w, notFound, http.StatusNotFound, nil)
			return
		}
		render(w, tool, http.StatusOK, map[string]any{"Tool": t, "Title": t.Name})
	})
	mux.HandleFunc("GET /guides/{slug}", func(w http.ResponseWriter, r *http.Request) {
		for _, g := range guides {
			if g.Slug == r.PathValue("slug") {
				render(w, guidePages[g.Slug], http.StatusOK, map[string]any{"Title": g.Title})
				return
			}
		}
		render(w, notFound, http.StatusNotFound, nil)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, notFound, http.StatusNotFound, nil)
	})
	return securityHeaders(dropWWW(mux))
}

// dropWWW sends www.example.com/anything to example.com/anything, so the site
// has one address. Caddy passes the Host header through unchanged.
func dropWWW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if host, ok := strings.CutPrefix(r.Host, "www."); ok {
			http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// render executes t into a buffer first, so a template error becomes a 500
// instead of half a page.
func render(w http.ResponseWriter, t *template.Template, status int, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		slog.Error("rendering page", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// securityHeaders sets headers every response gets. The site loads nothing
// from other origins, so the content security policy can be strict.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self'; style-src 'self'; font-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func cacheFor(d time.Duration, next http.Handler) http.Handler {
	maxAge := "public, max-age=" + strconv.Itoa(int(d.Seconds()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", maxAge)
		next.ServeHTTP(w, r)
	})
}

// listen adopts the socket systemd opened for this service (socket
// activation, as unit sets it up), or binds $PORT (default 8080) when run
// by hand.
func listen() (net.Listener, error) {
	if os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()) && os.Getenv("LISTEN_FDS") == "1" {
		return net.FileListener(os.NewFile(3, "systemd-socket"))
	}
	return net.Listen("tcp", ":"+envOr("PORT", "8080"))
}

// notifySystemd sends state to systemd's notify socket. It does nothing
// outside systemd.
func notifySystemd(state string) {
	socket := os.Getenv("NOTIFY_SOCKET")
	if socket == "" {
		return
	}
	conn, err := net.Dial("unixgram", socket)
	if err != nil {
		slog.Warn("systemd notify", "err", err)
		return
	}
	defer conn.Close()
	_, _ = conn.Write([]byte(state))
}

// watchdog pings systemd at half the interval it asked for.
func watchdog() {
	usec, err := strconv.Atoi(os.Getenv("WATCHDOG_USEC"))
	if err != nil || usec <= 0 {
		return
	}
	for {
		notifySystemd("WATCHDOG=1")
		time.Sleep(time.Duration(usec) * time.Microsecond / 2)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
