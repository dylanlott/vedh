package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/openmtg/edh-go/pkg/deckimport"
)

// allowAnyDialControl is a permissive ControlContext stand-in used by the
// tests below that need a real dial to a loopback httptest server —
// deliberately bypassing this file's own strict address/port denial
// (safeControl), which is proven separately and exhaustively by
// TestSafeControl_DeniedAddresses (direct calls, no socket at all) and
// TestSafeClient_RebindingFailsClosed (a real dial against the actual
// production safeControl). newSafeProviderClient — the zero-option
// constructor every production call site uses — always wires the real
// safeControl; nothing in this file ever changes that.
func allowAnyDialControl(context.Context, string, string, syscall.RawConn) error {
	return nil
}

// hermeticTestClient builds a client through the same construction path as
// production, with a permissive dial control (see allowAnyDialControl) and
// TLS certificate verification disabled so it can talk to a loopback
// httptest.NewTLSServer's self-signed certificate. Every other layer —
// CheckRedirect, the body cap, the timeouts — is the real, unmodified
// production logic.
func hermeticTestClient(allowedHosts map[string]struct{}, connectTimeout, totalTimeout time.Duration) *http.Client {
	c := newSafeProviderClientWithOptions(allowedHosts, connectTimeout, totalTimeout, allowAnyDialControl, nil)
	c.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // test-only: loopback self-signed cert
	return c
}

// TestSafeControl_DeniedAddresses calls safeControl directly, opening
// no socket at all, over the nine cases 01-06-PLAN.md names: the metadata
// address, a loopback address, a shared-address-space address, the
// metadata address written as an IPv4-mapped IPv6 literal, a
// protocol-assignment address, a benchmarking address, a public address on
// the wrong port, a public address on a non-TCP network, and one public
// address on the correct port that must be allowed.
func TestSafeControl_DeniedAddresses(t *testing.T) {
	cases := []struct {
		name    string
		network string
		address string
		wantErr bool
	}{
		{"metadata address", "tcp4", "169.254.169.254:443", true},
		{"loopback address", "tcp4", "127.0.0.1:443", true},
		{"shared address space", "tcp4", "100.64.0.1:443", true},
		{"metadata as IPv4-mapped IPv6", "tcp6", "[::ffff:169.254.169.254]:443", true},
		{"protocol assignment", "tcp4", "192.0.0.1:443", true},
		{"benchmarking", "tcp4", "198.18.0.1:443", true},
		{"public address, wrong port", "tcp4", "8.8.8.8:80", true},
		{"public address, non-TCP network", "udp4", "8.8.8.8:443", true},
		{"public address, correct port", "tcp4", "8.8.8.8:443", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := safeControl(context.Background(), tc.network, tc.address, nil)
			if tc.wantErr && err == nil {
				t.Fatalf("safeControl(%q, %q) = nil, want an error", tc.network, tc.address)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("safeControl(%q, %q) = %v, want nil", tc.network, tc.address, err)
			}
		})
	}
}

// TestSafeControl_ReportsFirstMatchingPrefix asserts firstMatchingDenyPrefix
// names the first prefix in declared order for an address inside two — a
// deliberately synthetic, overlapping pair, since the real deniedPrefixesV4/deniedPrefixesV6
// lists are disjoint IANA-registered blocks by construction. The property
// under test is generic to the ordering rule itself, not specific to
// today's production list.
func TestSafeControl_ReportsFirstMatchingPrefix(t *testing.T) {
	broad := netip.MustParsePrefix("10.0.0.0/8")
	narrow := netip.MustParsePrefix("10.0.0.0/16")
	addr := netip.MustParseAddr("10.0.1.1") // inside both broad and narrow

	if got, ok := firstMatchingDenyPrefix(addr, []netip.Prefix{broad, narrow}); !ok || got != broad {
		t.Fatalf("firstMatchingDenyPrefix() = (%v, %v), want (%v, true) when broad is declared first", got, ok, broad)
	}
	if got, ok := firstMatchingDenyPrefix(addr, []netip.Prefix{narrow, broad}); !ok || got != narrow {
		t.Fatalf("firstMatchingDenyPrefix() = (%v, %v), want (%v, true) when narrow is declared first", got, ok, narrow)
	}
}

// TestSafeClient_HostAndRedirect covers all six listed cases against real
// loopback TLS servers: the positive allowlisted case; a host differing
// only in letter case, accepted; a host containing an allowlisted host as
// a substring, refused; a redirect to an unallowlisted host, refused with
// its handler never invoked; a redirect to a non-secure scheme, refused;
// and a chain long enough to trip the hop bound.
func TestSafeClient_HostAndRedirect(t *testing.T) {
	allowedHosts := map[string]struct{}{"localhost": {}}
	client := hermeticTestClient(allowedHosts, 2*time.Second, 5*time.Second)

	t.Run("AllowlistedHostAccepted", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		resp, err := client.Get("https://localhost:" + portOf(t, srv) + "/")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("CaseInsensitiveHostAccepted", func(t *testing.T) {
		var srv *httptest.Server
		srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/target" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, "https://LOCALHOST:"+portOf(t, srv)+"/target", http.StatusFound)
		}))
		defer srv.Close()

		resp, err := client.Get("https://localhost:" + portOf(t, srv) + "/")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("SuffixHostRefused", func(t *testing.T) {
		var srv *httptest.Server
		srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://notlocalhost:"+portOf(t, srv)+"/target", http.StatusFound)
		}))
		defer srv.Close()

		_, err := client.Get("https://localhost:" + portOf(t, srv) + "/")
		if err == nil {
			t.Fatal("expected an error for a redirect host that only contains the allowlisted host as a substring")
		}
		if !strings.Contains(err.Error(), "blocked redirect host") {
			t.Fatalf("error = %v, want a blocked-redirect-host error", err)
		}
	})

	t.Run("RedirectToUnallowlistedHostNeverInvoked", func(t *testing.T) {
		var invoked atomic.Bool
		target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			invoked.Store(true)
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// target.URL is "https://127.0.0.1:<port>" -- a distinct host
			// string from the allowlisted "localhost", so this hop must be
			// refused before target's listener ever accepts a connection.
			http.Redirect(w, r, target.URL+"/", http.StatusFound)
		}))
		defer srv.Close()

		_, err := client.Get("https://localhost:" + portOf(t, srv) + "/")
		if err == nil {
			t.Fatal("expected an error for a redirect to an unallowlisted host")
		}
		if !strings.Contains(err.Error(), "blocked redirect host") {
			t.Fatalf("error = %v, want a blocked-redirect-host error", err)
		}
		if invoked.Load() {
			t.Fatal("the unallowlisted target's handler was invoked; it must never be reached")
		}
	})

	t.Run("NonSecureSchemeRefused", func(t *testing.T) {
		var srv *httptest.Server
		srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://localhost:"+portOf(t, srv)+"/target", http.StatusFound)
		}))
		defer srv.Close()

		_, err := client.Get("https://localhost:" + portOf(t, srv) + "/")
		if err == nil {
			t.Fatal("expected an error for a redirect to a non-secure scheme")
		}
		if !strings.Contains(err.Error(), "blocked redirect scheme") {
			t.Fatalf("error = %v, want a blocked-redirect-scheme error", err)
		}
	})

	t.Run("HopBoundTripped", func(t *testing.T) {
		var srv *httptest.Server
		srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Redirects to itself unconditionally: two hops must be
			// followed and the third refused.
			http.Redirect(w, r, "https://localhost:"+portOf(t, srv)+"/next", http.StatusFound)
		}))
		defer srv.Close()

		_, err := client.Get("https://localhost:" + portOf(t, srv) + "/")
		if err == nil {
			t.Fatal("expected an error once the redirect chain exceeds the hop bound")
		}
		if !strings.Contains(err.Error(), "too many redirects") {
			t.Fatalf("error = %v, want a too-many-redirects error", err)
		}
	})
}

// TestSafeClient_InitialRequestHostEnforced is a regression test for
// WR-01: the host allowlist must be enforced on the INITIAL request, not
// only on redirect targets. Before the fix, fetchDeckProviderURL validated
// only the scheme and that the host was non-empty; DECK_PROVIDER_ALLOWED_
// HOSTS was consulted only inside newDeckProviderCheckRedirect, which Go's
// http.Client invokes solely before following a redirect hop. This proves
// a host absent from allowedHosts is refused before any dial is attempted
// -- the target server's handler is never invoked -- exactly the same
// zero-dial-attempt proof this file's kill-switch tests already use for
// deckProviderFetch, applied one layer down to fetchDeckProviderURL itself.
func TestSafeClient_InitialRequestHostEnforced(t *testing.T) {
	var invoked atomic.Bool
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		invoked.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	client := hermeticTestClient(nil, 2*time.Second, 5*time.Second)

	// Deliberately an allowlist that does NOT contain target's host, so
	// the initial-request check must refuse the dial on its own -- there
	// is no redirect involved here at all.
	_, err := fetchDeckProviderURL(context.Background(), client, map[string]struct{}{"totally-unrelated.example": {}}, target.URL)
	if err == nil {
		t.Fatal("expected an error for a host absent from the allowlist on the initial request")
	}
	if !strings.Contains(err.Error(), "blocked host") {
		t.Fatalf("error = %v, want a blocked-host error", err)
	}
	if invoked.Load() {
		t.Fatal("the unallowlisted target's handler was invoked; the initial request must never dial it")
	}
}

// TestSafeClient_RejectsBeforeResolution asserts an empty host and a
// non-secure scheme are refused by fetchDeckProviderURL before any name
// resolution is attempted, using a resolver stub that records whether it
// was ever invoked.
func TestSafeClient_RejectsBeforeResolution(t *testing.T) {
	var resolverCalled atomic.Bool
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			resolverCalled.Store(true)
			return nil, errors.New("resolver dial invoked unexpectedly")
		},
	}
	client := newSafeProviderClientWithOptions(nil, 2*time.Second, 5*time.Second, safeControl, resolver)

	if _, err := fetchDeckProviderURL(context.Background(), client, nil, "https://"); err == nil {
		t.Fatal("expected an error for an empty host")
	}
	if resolverCalled.Load() {
		t.Fatal("expected no name resolution for an empty host")
	}

	// ".invalid" is IANA-reserved (RFC 2606) to never resolve; naming it
	// here is itself part of the proof that no resolution is attempted --
	// unlike a real domain, there is no ambiguity about whether this host
	// could ever be "reached".
	if _, err := fetchDeckProviderURL(context.Background(), client, nil, "http://deck-provider.invalid/deck"); err == nil {
		t.Fatal("expected an error for a non-secure scheme")
	}
	if resolverCalled.Load() {
		t.Fatal("expected no name resolution for a non-secure scheme")
	}
}

// TestSafeClient_RebindingFailsClosed proves safeControl is actually
// wired into the real production client (newSafeProviderClient, the
// zero-option constructor every production call site uses), not merely
// correct in isolation: a request to a hostname that resolves to loopback
// is refused at dial time. Per 01-RESEARCH.md Validation Architecture
// subsection 4's own stated fallback, this is the "hostname resolving to
// loopback fails" equivalent of a full DNS-rebinding harness, and needs no
// stub DNS server: ControlContext runs after resolution and before
// connect(2), so the request never leaves the process regardless of
// whether an actual listener exists at that address.
func TestSafeClient_RebindingFailsClosed(t *testing.T) {
	client := newSafeProviderClient(map[string]struct{}{"localhost": {}})

	_, err := client.Get("https://localhost/")
	if err == nil {
		t.Fatal("expected a request to a hostname resolving to loopback to be refused")
	}
	if !strings.Contains(err.Error(), "blocked address") {
		t.Fatalf("error = %v, want a blocked-address error from safeControl", err)
	}
}

// TestSafeClient_BodyCap proves the response cap is enforced on the
// counted length actually read, never on a declared length: exactly the
// cap is accepted, the cap plus one byte is refused, and a response with
// no declared length at all (forced chunked transfer, so
// resp.ContentLength reports -1/unknown) carrying far more than the cap is
// also refused. Go's client-side Content-Length framing makes a
// *declared-small-but-actually-larger* body structurally unreproducible
// against a standards-compliant client (the framing itself bounds the
// read to the declared length) — the chunked/undeclared-length case below
// is the actual mechanism by which a response can exceed what
// resp.ContentLength would suggest, which is exactly why this
// implementation never consults it at all.
func TestSafeClient_BodyCap(t *testing.T) {
	client := hermeticTestClient(nil, 2*time.Second, 5*time.Second)

	t.Run("ExactlyCapAccepted", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(make([]byte, deckProviderMaxBodyBytes))
		}))
		defer srv.Close()

		data, err := fetchDeckProviderURL(context.Background(), client, allowedHostsForServer(t, srv), srv.URL)
		if err != nil {
			t.Fatalf("fetchDeckProviderURL() error = %v", err)
		}
		if len(data) != deckProviderMaxBodyBytes {
			t.Fatalf("len(data) = %d, want %d", len(data), deckProviderMaxBodyBytes)
		}
	})

	t.Run("OverCapRefused", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(make([]byte, deckProviderMaxBodyBytes+1))
		}))
		defer srv.Close()

		_, err := fetchDeckProviderURL(context.Background(), client, allowedHostsForServer(t, srv), srv.URL)
		if err == nil {
			t.Fatal("expected an error for a body one byte over the cap")
		}
		if !strings.Contains(err.Error(), "1 MiB cap") {
			t.Fatalf("error = %v, want a 1 MiB cap error", err)
		}
	})

	t.Run("ChunkedUndeclaredLengthOverCapRefused", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			flusher := w.(http.Flusher)
			buf := make([]byte, 4096)
			written := 0
			for written < 2*deckProviderMaxBodyBytes {
				n, werr := w.Write(buf)
				written += n
				flusher.Flush()
				if werr != nil {
					return
				}
			}
		}))
		defer srv.Close()

		_, err := fetchDeckProviderURL(context.Background(), client, allowedHostsForServer(t, srv), srv.URL)
		if err == nil {
			t.Fatal("expected an error for a chunked, undeclared-length body over the cap")
		}
		if !strings.Contains(err.Error(), "1 MiB cap") {
			t.Fatalf("error = %v, want a 1 MiB cap error", err)
		}
	})
}

// TestSafeClient_Timeouts asserts a bounded elapsed time and a timeout
// error for both the connect and the total budget, with both budgets
// parameterised down so the test stays fast.
func TestSafeClient_Timeouts(t *testing.T) {
	t.Run("ConnectTimeout", func(t *testing.T) {
		const budget = 200 * time.Millisecond
		client := newSafeProviderClientWithOptions(nil, budget, 5*time.Second, allowAnyDialControl, nil)

		start := time.Now()
		// 10.255.255.1 is a private, non-routable RFC 1918 address that
		// answers no packets at all -- the standard deterministic
		// black-hole target for a connect-timeout test. This is not "a
		// live deck provider": nothing is ever served, read, or depended
		// on from it, and allowAnyDialControl (not the production
		// safeControl, which would refuse a 10.0.0.0/8 literal
		// outright) is what's actually under test here -- the connect
		// budget, not the address deny list, which TestSafeControl_
		// DeniedAddresses already covers hermetically with no socket at
		// all.
		_, err := client.Get("https://10.255.255.1/")
		elapsed := time.Since(start)

		if err == nil {
			t.Fatal("expected the connect attempt to fail")
		}
		if elapsed > 2*time.Second {
			t.Fatalf("elapsed = %v, want well under 2s given a %v connect budget", elapsed, budget)
		}
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("error = %v, want a timeout error", err)
		}
	})

	t.Run("TotalTimeout", func(t *testing.T) {
		const budget = 150 * time.Millisecond
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(600 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := hermeticTestClient(nil, 5*time.Second, budget)

		start := time.Now()
		_, err := client.Get(srv.URL)
		elapsed := time.Since(start)

		if err == nil {
			t.Fatal("expected the request to fail once the total budget elapses")
		}
		if elapsed > 2*time.Second {
			t.Fatalf("elapsed = %v, want well under 2s given a %v total budget", elapsed, budget)
		}
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("error = %v, want a timeout error", err)
		}
	})
}

// TestSafeClient_ZeroByteBodyIsProviderError asserts an empty response
// body yields a normalized provider error rather than an empty (but
// "successful") result.
func TestSafeClient_ZeroByteBodyIsProviderError(t *testing.T) {
	client := hermeticTestClient(nil, 2*time.Second, 5*time.Second)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := fetchDeckProviderURL(context.Background(), client, allowedHostsForServer(t, srv), srv.URL)
	if err == nil {
		t.Fatal("expected an error for a zero-byte response body")
	}
	if !strings.Contains(err.Error(), "empty response body") {
		t.Fatalf("error = %v, want an empty-response-body error", err)
	}
}

// portOf returns srv's loopback port as a string, failing the test if srv's
// URL cannot be parsed.
func portOf(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	_, port, err := net.SplitHostPort(strings.TrimPrefix(strings.TrimPrefix(srv.URL, "https://"), "http://"))
	if err != nil {
		t.Fatalf("portOf(%q): %v", srv.URL, err)
	}
	return port
}

// allowedHostsForServer returns an allowlist containing exactly srv's
// loopback host, for tests that call fetchDeckProviderURL directly (WR-01
// fix, code review phase 01): fetchDeckProviderURL now enforces the host
// allowlist on the initial request, so any hermetic test dialing a real
// loopback httptest server needs that server's own host allowlisted or
// every such call would be refused before the behavior under test (body
// cap, timeouts, empty-body handling) is ever exercised.
func allowedHostsForServer(t *testing.T, srv *httptest.Server) map[string]struct{} {
	t.Helper()
	host, _, err := net.SplitHostPort(strings.TrimPrefix(strings.TrimPrefix(srv.URL, "https://"), "http://"))
	if err != nil {
		t.Fatalf("allowedHostsForServer(%q): %v", srv.URL, err)
	}
	return map[string]struct{}{strings.ToLower(host): {}}
}

// withDeckProviderFetchSpy substitutes deckProviderFetch with a spy that
// increments called, restoring the real fetchDeckProviderURL via
// t.Cleanup. Every kill-switch test uses this to prove zero dial attempts
// were made -- not merely that an error was returned -- since a call this
// resolver never makes is a call this spy never counts.
func withDeckProviderFetchSpy(t *testing.T) *int {
	t.Helper()
	called := 0
	original := deckProviderFetch
	deckProviderFetch = func(ctx context.Context, client *http.Client, allowedHosts map[string]struct{}, rawURL string) ([]byte, error) {
		called++
		return nil, nil
	}
	t.Cleanup(func() { deckProviderFetch = original })
	return &called
}

// TestProvider_KillSwitch asserts that with the flag unset, a deck URL
// yields exactly one blocking error mentioning pasting text, CanContinue
// is false, and deckProviderFetch is never called -- proving zero dial
// attempts, not merely that an error was returned.
func TestProvider_KillSwitch(t *testing.T) {
	s := testAPI(t)
	s.cfg.DeckProviderEnabled = false
	s.deckProviderAllowedHosts = map[string]struct{}{archidektHost: {}}
	called := withDeckProviderFetchSpy(t)

	rawURL := "https://" + archidektHost + "/api/decks/13074677/"
	input := InputDeckImport{SourceURL: &rawURL, SessionID: "kill-switch-test"}

	preview, err := s.PreviewDeck(context.Background(), input)
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if preview.CanContinue {
		t.Fatal("expected CanContinue = false with the provider flag unset")
	}
	if len(preview.BlockingErrors) != 1 {
		t.Fatalf("expected exactly one blocking error, got %d: %v", len(preview.BlockingErrors), preview.BlockingErrors)
	}
	if !strings.Contains(strings.ToLower(preview.BlockingErrors[0]), "paste") {
		t.Fatalf("blocking error %q does not mention pasting text as a fallback", preview.BlockingErrors[0])
	}
	if *called != 0 {
		t.Fatalf("deckProviderFetch was called %d times, want 0", *called)
	}
}

// TestProvider_KillSwitchRequiresAllowlist asserts the flag set with an
// empty allowlist is treated as disabled, identically to the flag being
// unset: still no dial attempt.
func TestProvider_KillSwitchRequiresAllowlist(t *testing.T) {
	s := testAPI(t)
	s.cfg.DeckProviderEnabled = true
	s.deckProviderAllowedHosts = map[string]struct{}{}
	called := withDeckProviderFetchSpy(t)

	if s.providerEnabled() {
		t.Fatal("expected providerEnabled() = false when the allowlist is empty, even with the flag set")
	}

	rawURL := "https://" + archidektHost + "/api/decks/13074677/"
	input := InputDeckImport{SourceURL: &rawURL, SessionID: "kill-switch-allowlist-test"}

	preview, err := s.PreviewDeck(context.Background(), input)
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if preview.CanContinue {
		t.Fatal("expected CanContinue = false with an empty allowlist")
	}
	if *called != 0 {
		t.Fatalf("deckProviderFetch was called %d times, want 0", *called)
	}
}

// TestProvider_PasteUnaffectedByFlag asserts a text-only preview produces
// a byte-equal result whether the provider flag is set or unset -- the
// pasted-text path must be entirely unaffected by the flag in either
// state.
func TestProvider_PasteUnaffectedByFlag(t *testing.T) {
	text := "1 Sol Ring"

	run := func(t *testing.T, enabled bool) []byte {
		s := testAPI(t)
		s.cfg.DeckProviderEnabled = enabled
		if enabled {
			s.deckProviderAllowedHosts = map[string]struct{}{archidektHost: {}}
		} else {
			s.deckProviderAllowedHosts = map[string]struct{}{}
		}

		input := InputDeckImport{Text: &text, SessionID: "paste-unaffected-test"}
		preview, err := s.PreviewDeck(context.Background(), input)
		if err != nil {
			t.Fatalf("PreviewDeck() error = %v", err)
		}
		data, err := json.Marshal(preview)
		if err != nil {
			t.Fatalf("json.Marshal(preview) error = %v", err)
		}
		return data
	}

	disabled := run(t, false)
	enabled := run(t, true)

	if string(disabled) != string(enabled) {
		t.Fatalf("preview differs by provider flag state:\ndisabled: %s\nenabled:  %s", disabled, enabled)
	}
}

// -----------------------------------------------------------------------
// D-14 reversed (plan 01-08, gap G-01-1): Archidekt routing, Moxfield
// scaffold removed
// -----------------------------------------------------------------------
//
// Every test below is hermetic: deckProviderFetch is either spied (never
// the real fetchDeckProviderURL) or the adapter's normalizer is called
// directly with an in-memory byte slice. None makes an outbound
// connection. archidektAdapter's field-mapping contract tests (against
// the committed fixture) are plan 01-09's job, not this one -- the tests
// below prove the routing gate itself, which is this plan's scope.

// TestProvider_DeckProviderAdapterFor proves the hostname lookup is
// case-insensitive, matches only the registered Archidekt host, and
// treats an unparseable URL, an empty host, or a host with no registered
// adapter (Moxfield, since G-01-1) as "no adapter" rather than panicking
// or matching by accident.
func TestProvider_DeckProviderAdapterFor(t *testing.T) {
	if _, ok := deckProviderAdapterFor("https://" + archidektHost + "/api/decks/13074677/"); !ok {
		t.Fatalf("expected an adapter registered for %q", archidektHost)
	}
	if _, ok := deckProviderAdapterFor("https://ARCHIDEKT.COM/api/decks/13074677/"); !ok {
		t.Fatal("expected the hostname lookup to be case-insensitive")
	}
	if _, ok := deckProviderAdapterFor("https://moxfield.com/decks/abc123"); ok {
		t.Fatal("expected no adapter registered for moxfield.com (reversed at G-01-1)")
	}
	if _, ok := deckProviderAdapterFor("not a url at all"); ok {
		t.Fatal("expected no adapter for an unparseable URL")
	}
	if _, ok := deckProviderAdapterFor("https:///decks/1"); ok {
		t.Fatal("expected no adapter for an empty host")
	}
}

// TestProvider_ArchidektAdapterRoutesToRealNormalizer proves the routing
// gate genuinely reaches the secure fetch client (the spy is called
// exactly once) and genuinely reaches archidektAdapter's real normalizer:
// handed the spy's zero-value nil body, normalizeToDeckText correctly
// fails closed on malformed JSON (its first parseable failure mode, not a
// forever-unimplemented seam the way moxfieldAdapter's was before G-01-1).
// This is the proof that previewDeckURL's routing decision is real, not
// merely that the end result happens to be a blocking error.
func TestProvider_ArchidektAdapterRoutesToRealNormalizer(t *testing.T) {
	s := testAPI(t)
	s.cfg.DeckProviderEnabled = true
	s.deckProviderAllowedHosts = map[string]struct{}{archidektHost: {}}
	called := withDeckProviderFetchSpy(t)

	rawURL := "https://" + archidektHost + "/api/decks/13074677/"
	input := InputDeckImport{SourceURL: &rawURL, SessionID: "archidekt-routing-test"}

	preview, err := s.PreviewDeck(context.Background(), input)
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if preview.CanContinue {
		t.Fatal("expected CanContinue = false: the spy's nil body is not valid JSON")
	}
	if len(preview.BlockingErrors) != 1 {
		t.Fatalf("expected exactly one blocking error, got %d: %v", len(preview.BlockingErrors), preview.BlockingErrors)
	}
	if !strings.Contains(strings.ToLower(preview.BlockingErrors[0]), "paste") {
		t.Fatalf("blocking error %q does not mention pasting text as a fallback", preview.BlockingErrors[0])
	}
	if *called != 1 {
		t.Fatalf("deckProviderFetch was called %d times, want exactly 1 -- the routing gate should genuinely reach the secure client", *called)
	}
}

// TestProvider_ArchidektAdapterMalformedBody proves
// archidektAdapter.normalizeToDeckText, called directly, returns the
// static errArchidektMalformedResponse sentinel for a nil or empty body
// and errArchidektNoCardRows for well-formed JSON with zero card rows --
// neither sentinel ever carries the input bytes (server/deck_import.go's
// error-hygiene comment).
func TestProvider_ArchidektAdapterMalformedBody(t *testing.T) {
	adapter := archidektAdapter{}

	if got := adapter.source(); got != deckimport.SourceArchidekt {
		t.Fatalf("source() = %q, want %q", got, deckimport.SourceArchidekt)
	}

	malformed := [][]byte{nil, []byte(""), []byte("not json")}
	for _, body := range malformed {
		text, err := adapter.normalizeToDeckText(body)
		if !errors.Is(err, errArchidektMalformedResponse) {
			t.Fatalf("normalizeToDeckText(%q) error = %v, want errArchidektMalformedResponse", body, err)
		}
		if text != "" {
			t.Fatalf("normalizeToDeckText(%q) text = %q, want empty on error", body, text)
		}
	}

	text, err := adapter.normalizeToDeckText([]byte(`{"categories":[],"cards":[]}`))
	if !errors.Is(err, errArchidektNoCardRows) {
		t.Fatalf("normalizeToDeckText(no cards) error = %v, want errArchidektNoCardRows", err)
	}
	if text != "" {
		t.Fatalf("normalizeToDeckText(no cards) text = %q, want empty on error", text)
	}
}

// TestProvider_UnregisteredHostNeverFetched proves a host that is
// allowlisted (so DECK_PROVIDER_ALLOWED_HOSTS alone would not refuse it)
// but has no registered adapter is still refused before any fetch is
// attempted: the allowlist and the adapter registry are independent
// gates, and both must agree before a dial is ever made. moxfield.com is
// the deliberately unregistered host here, post-G-01-1.
func TestProvider_UnregisteredHostNeverFetched(t *testing.T) {
	s := testAPI(t)
	s.cfg.DeckProviderEnabled = true
	s.deckProviderAllowedHosts = map[string]struct{}{archidektHost: {}, "moxfield.com": {}}
	called := withDeckProviderFetchSpy(t)

	rawURL := "https://moxfield.com/decks/abc123"
	input := InputDeckImport{SourceURL: &rawURL, SessionID: "unregistered-host-test"}

	preview, err := s.PreviewDeck(context.Background(), input)
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if preview.CanContinue {
		t.Fatal("expected CanContinue = false for a host with no registered adapter")
	}
	if *called != 0 {
		t.Fatalf("deckProviderFetch was called %d times, want 0: no adapter is registered for moxfield.com even though it is allowlisted", *called)
	}
}

// TestProvider_KillSwitchStableAcrossAdapterChange re-proves plan 01-06's
// TestProvider_KillSwitch/TestProvider_PasteUnaffectedByFlag properties
// against the now-registered Archidekt host: with the kill switch off, an
// archidekt.com deck URL still returns the ordinary paste-fallback error
// and never reaches deckProviderFetch, and the pasted-text path is
// entirely unaffected by which provider is registered.
func TestProvider_KillSwitchStableAcrossAdapterChange(t *testing.T) {
	t.Run("KillSwitchOffNeverFetches", func(t *testing.T) {
		s := testAPI(t)
		s.cfg.DeckProviderEnabled = false
		s.deckProviderAllowedHosts = map[string]struct{}{archidektHost: {}}
		called := withDeckProviderFetchSpy(t)

		rawURL := "https://" + archidektHost + "/api/decks/13074677/"
		input := InputDeckImport{SourceURL: &rawURL, SessionID: "archidekt-stability-off-test"}

		preview, err := s.PreviewDeck(context.Background(), input)
		if err != nil {
			t.Fatalf("PreviewDeck() error = %v", err)
		}
		if preview.CanContinue {
			t.Fatal("expected CanContinue = false with the kill switch off")
		}
		if *called != 0 {
			t.Fatalf("deckProviderFetch was called %d times, want 0 with the kill switch off", *called)
		}
	})

	t.Run("PasteUnaffected", func(t *testing.T) {
		text := "1 Sol Ring"
		s := testAPI(t)
		s.cfg.DeckProviderEnabled = true
		s.deckProviderAllowedHosts = map[string]struct{}{archidektHost: {}}

		input := InputDeckImport{Text: &text, SessionID: "archidekt-stability-paste-test"}
		preview, err := s.PreviewDeck(context.Background(), input)
		if err != nil {
			t.Fatalf("PreviewDeck() error = %v", err)
		}
		if !preview.CanContinue {
			t.Fatal("expected a pasted decklist to resolve normally regardless of which provider is registered")
		}
	})
}
