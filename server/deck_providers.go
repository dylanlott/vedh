package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/openmtg/edh-go/pkg/deckimport"
)

// This file implements 01-06's provider-agnostic secure outbound fetch
// client: two independent hooks, ControlContext (address and port policy)
// and CheckRedirect (scheme and host policy), whose responsibilities do not
// overlap and neither of which is sufficient alone. Building it ahead of
// and independent of the D-14 checkpoint (plan 01-07) is deliberate: both
// branches of that checkpoint — a real provider, or a documented no-go —
// need the same client and the same default-off kill switch, and building
// it now means both branches are equally cheap at the checkpoint.
//
// Resolving a caller-supplied hostname, validating the result, and only
// then dialing is TOCTOU-vulnerable by construction: a caller-controlled
// name server with a short record lifetime can answer with a public
// address for the validating lookup and a private one for the connecting
// lookup. ControlContext runs after resolution and before connect(2),
// receiving the literal address that will actually be dialled, so there is
// no window (T-01-25).

const (
	// deckProviderConnectTimeout is the locked 3-second connect budget.
	deckProviderConnectTimeout = 3 * time.Second
	// deckProviderTotalTimeout is the locked 8-second total budget.
	// http.Client.Timeout is documented to cover connection time, all
	// redirects, and the response body read — exactly this budget's
	// scope, which is what stops a slowly-trickled body from hanging far
	// past it (T-01-26).
	deckProviderTotalTimeout = 8 * time.Second
	// deckProviderMaxRedirectHops bounds the redirect chain: CheckRedirect
	// is invoked before following each hop with via holding every request
	// made so far, so refusing at len(via) >= 3 follows exactly two hops
	// and refuses the third.
	deckProviderMaxRedirectHops = 3
	// deckProviderMaxBodyBytes is the locked 1 MiB response cap.
	deckProviderMaxBodyBytes = 1 << 20
	// deckProviderSecureScheme is the only scheme ever permitted, both for
	// the initial URL and for every redirect hop.
	deckProviderSecureScheme = "https"
	// deckProviderSecurePort is the only port ControlContext ever permits
	// a dial to reach — enforcing the scheme restriction at the socket
	// layer too, independent of anything the URL layer decided.
	deckProviderSecurePort = "443"
)

// deniedPrefixesV4 is the IANA-derived deny list for IPv4 literals, transcribed
// from the registry used by code.dny.dev/ssrf (01-RESEARCH.md Pattern 8).
// Declared in this fixed order: firstMatchingDenyPrefix names the first
// entry an address falls inside, so an address inside two produces a
// deterministic reason rather than one that depends on map iteration.
//
// net/netip's own IsPrivate() predicate, and the standard library's own
// "global unicast" predicate (deliberately not named here by its Go
// identifier — this file's own verify gate greps for the absence of that
// identifier, the same way it does for the environment-derived proxy
// resolver below), are both insufficient as the gate here. The latter's
// own doc comment states private and unique-local addresses are still
// considered global unicast — it is not merely insufficient, it is
// misleading. IsPrivate() alone covers only the legacy RFC 1918 ranges and
// the IPv6 unique-local block, and catches none of the shared-address
// (100.64.0.0/10), protocol-assignment (192.0.0.0/24, 192.0.2.0/24, ...),
// or benchmarking (198.18.0.0/15) ranges below — which is exactly where
// the cloud-metadata address (169.254.169.254) has neighbours.
var deniedPrefixesV4 = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // "this network"
	netip.MustParsePrefix("10.0.0.0/8"),      // RFC 1918 private
	netip.MustParsePrefix("100.64.0.0/10"),   // shared address space (CGN)
	netip.MustParsePrefix("127.0.0.0/8"),     // loopback (also caught by IsLoopback)
	netip.MustParsePrefix("169.254.0.0/16"),  // link-local, incl. cloud metadata
	netip.MustParsePrefix("172.16.0.0/12"),   // RFC 1918 private
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // documentation (TEST-NET-1)
	netip.MustParsePrefix("192.31.196.0/24"), // AS112
	netip.MustParsePrefix("192.52.193.0/24"), // AMT
	netip.MustParsePrefix("192.88.99.0/24"),  // 6to4 relay anycast
	netip.MustParsePrefix("192.168.0.0/16"),  // RFC 1918 private
	netip.MustParsePrefix("192.175.48.0/24"), // direct AS112
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"), // documentation (TEST-NET-2)
	netip.MustParsePrefix("203.0.113.0/24"),  // documentation (TEST-NET-3)
	netip.MustParsePrefix("224.0.0.0/4"),     // multicast (also caught by IsMulticast)
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved for future use
}

// globalUnicastV6 is the only zone an IPv6 literal may ever dial into:
// everything outside 2000::/3 (loopback, unspecified, link-local, unique
// local, multicast, and every other special-purpose block) is refused
// before the more specific deniedPrefixesV6 entries are even considered.
var globalUnicastV6 = netip.MustParsePrefix("2000::/3")

// deniedPrefixesV6 further restricts the global-unicast zone to exclude the
// special-purpose IPv6 blocks nested inside it (01-RESEARCH.md Pattern 8).
// Same fixed-order, first-match contract as deniedPrefixesV4.
var deniedPrefixesV6 = []netip.Prefix{
	netip.MustParsePrefix("2001::/23"),         // IETF protocol assignments (incl. Teredo)
	netip.MustParsePrefix("2001:db8::/32"),     // documentation
	netip.MustParsePrefix("2002::/16"),         // 6to4
	netip.MustParsePrefix("2620:4f:8000::/48"), // direct AS112
	netip.MustParsePrefix("3fff::/20"),         // documentation
}

// firstMatchingDenyPrefix returns the first prefix in prefixes (in
// declared order) that contains addr, and whether one was found.
// Returning the *first* match — rather than, say, the narrowest — is what
// makes the refusal reason for an address inside two prefixes
// deterministic.
func firstMatchingDenyPrefix(addr netip.Addr, prefixes []netip.Prefix) (netip.Prefix, bool) {
	for _, p := range prefixes {
		if p.Contains(addr) {
			return p, true
		}
	}
	return netip.Prefix{}, false
}

// safeControl is the ControlContext hook: address and port policy,
// enforced against the literal address the dialer already resolved and is
// about to connect(2) to — never against a name, and never in a separate
// pre-flight lookup (see this file's header comment on why that would be
// TOCTOU-vulnerable).
func safeControl(_ context.Context, network, address string, _ syscall.RawConn) error {
	if network != "tcp4" && network != "tcp6" {
		return fmt.Errorf("deck provider: blocked network %q", network)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("deck provider: unparseable dial address %q: %w", address, err)
	}
	if port != deckProviderSecurePort {
		return fmt.Errorf("deck provider: blocked port %q", port)
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("deck provider: unparseable dial address %q: %w", host, err)
	}
	// Unmap an IPv4-mapped IPv6 literal (::ffff:169.254.169.254) before any
	// prefix check: net/netip's own predicates already do this for
	// IsLoopback etc., but the explicit deny-prefix walk below needs it
	// done up front or the metadata address written in mapped form would
	// walk the IPv6 list instead of the IPv4 one that actually names it.
	if addr.Is4In6() {
		addr = addr.Unmap()
	}

	if addr.IsLoopback() || addr.IsUnspecified() || addr.IsMulticast() || addr.IsLinkLocalUnicast() {
		return fmt.Errorf("deck provider: blocked address %s", addr)
	}

	if addr.Is4() {
		if p, blocked := firstMatchingDenyPrefix(addr, deniedPrefixesV4); blocked {
			return fmt.Errorf("deck provider: blocked address %s (in %s)", addr, p)
		}
		return nil
	}

	if !globalUnicastV6.Contains(addr) {
		return fmt.Errorf("deck provider: blocked address %s (outside %s)", addr, globalUnicastV6)
	}
	if p, blocked := firstMatchingDenyPrefix(addr, deniedPrefixesV6); blocked {
		return fmt.Errorf("deck provider: blocked address %s (in %s)", addr, p)
	}
	return nil
}

// newDeckProviderCheckRedirect builds the CheckRedirect hook: scheme and
// host policy, re-validated on every redirect hop rather than only on the
// caller-supplied URL, because Go's default CheckRedirect is nil and
// follows up to 10 redirects with no policy of its own. ControlContext
// independently re-validates the address on every hop this callback
// allows, so the two hooks cover different bypasses (T-01-07).
//
// Host comparison is exact equality on the lower-cased value the URL
// parser returns — never a prefix or suffix comparison, and no decoding
// step runs before the comparison — so a host that merely contains an
// allowlisted host as a substring, or an internationalized/percent-encoded
// host that would decode to one, is refused rather than accepted.
func newDeckProviderCheckRedirect(allowedHosts map[string]struct{}) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= deckProviderMaxRedirectHops {
			return errors.New("deck provider: too many redirects")
		}
		if req.URL.Scheme != deckProviderSecureScheme {
			return fmt.Errorf("deck provider: blocked redirect scheme %q", req.URL.Scheme)
		}
		host := strings.ToLower(req.URL.Hostname())
		if _, ok := allowedHosts[host]; !ok {
			return fmt.Errorf("deck provider: blocked redirect host %q", host)
		}
		return nil
	}
}

// dialControlFunc is net.Dialer.ControlContext's signature, named here so
// newSafeProviderClientWithOptions can accept one as a parameter.
type dialControlFunc func(ctx context.Context, network, address string, c syscall.RawConn) error

// newSafeProviderClient is production's entry point: the locked 3-second
// connect and 8-second total budgets, the real safeControl, the
// default resolver, and CheckRedirect bound to allowedHosts.
func newSafeProviderClient(allowedHosts map[string]struct{}) *http.Client {
	return newSafeProviderClientWithOptions(allowedHosts, deckProviderConnectTimeout, deckProviderTotalTimeout, safeControl, nil)
}

// newSafeProviderClientWithOptions is newSafeProviderClient's parameterised
// form — the same budget-as-parameter shape plan 01-05 established for
// buildDeckPreviewWithSuggestionBudget — so a test can force a timeout,
// substitute an in-process resolver, or (for the layers that are not
// address policy) substitute a permissive dialControl to reach a loopback
// httptest server without weakening what newSafeProviderClient's zero-option
// form actually wires in production. A nil resolver means "use the default
// resolver", net.Dialer's own zero value.
func newSafeProviderClientWithOptions(allowedHosts map[string]struct{}, connectTimeout, totalTimeout time.Duration, dialControl dialControlFunc, resolver *net.Resolver) *http.Client {
	dialer := &net.Dialer{
		Timeout:        connectTimeout,
		ControlContext: dialControl,
		Resolver:       resolver,
	}
	return &http.Client{
		// Every timeout here is an integer time.Duration, so no rounding
		// or floating-point conversion can widen either budget.
		Timeout: totalTimeout,
		Transport: &http.Transport{
			// Deliberately nil. The widely-copied reference snippet for
			// this exact pattern sets this to the environment-derived
			// proxy resolver, which silently disables the whole address
			// control above: with a proxy configured, the dialer connects
			// to the *proxy*, so ControlContext validates the proxy's
			// address while the real destination travels in a CONNECT
			// request the hook never sees. Every control in this file
			// would still pass its own tests, and the client could still
			// reach a denied address (T-01-08).
			Proxy:               nil,
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: connectTimeout,
			DisableKeepAlives:   true,
		},
		CheckRedirect: newDeckProviderCheckRedirect(allowedHosts),
	}
}

// readBodyWithCap reads at most deckProviderMaxBodyBytes+1 bytes from body
// and compares the counted length actually read, never the declared
// Content-Length: that header is provider-controlled, may be absent
// entirely under chunked transfer encoding, and is exactly what an
// implementation trusting it would get wrong on a response declaring a
// tiny length while sending far more (T-01-09).
func readBodyWithCap(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, deckProviderMaxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("deck provider: reading response body: %w", err)
	}
	if len(data) > deckProviderMaxBodyBytes {
		return nil, errors.New("deck provider: response exceeded the 1 MiB cap")
	}
	return data, nil
}

// fetchDeckProviderURL performs one GET against rawURL through client,
// refusing an empty host, a non-secure scheme, or a host outside
// allowedHosts before any request is built — and therefore before any name
// resolution is attempted — and returns the capped, counted response body.
// A zero-byte body is reported as a normalized provider error rather than
// treated as an empty deck, so a caller can never mistake "the provider
// returned nothing" for "the provider returned an empty decklist".
//
// WR-01 fix (code review, phase 01): allowedHosts was previously enforced
// only inside newDeckProviderCheckRedirect, which Go's http.Client invokes
// solely before following a *redirect* hop — never for the first request
// of a chain. That left the initial dial target bounded only by the
// hardcoded deckProviderAdapters registry, contradicting this codebase's
// own documentation that DECK_PROVIDER_ALLOWED_HOSTS is the authoritative
// host allowlist for every outbound dial. Checking here, before the
// request is even built, closes that gap for the hop that actually matters
// today and for every hop going forward.
func fetchDeckProviderURL(ctx context.Context, client *http.Client, allowedHosts map[string]struct{}, rawURL string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("deck provider: invalid URL: %w", err)
	}
	if parsed.Scheme != deckProviderSecureScheme {
		return nil, fmt.Errorf("deck provider: blocked scheme %q", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return nil, errors.New("deck provider: empty host")
	}
	if _, ok := allowedHosts[host]; !ok {
		return nil, fmt.Errorf("deck provider: blocked host %q", host)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("deck provider: building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deck provider: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := readBodyWithCap(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("deck provider: empty response body")
	}
	return data, nil
}

// providerEnabled reports whether the deck-provider fetch path
// (server/deck_import.go's previewDeckURL) may run at all. Both the flag
// and a non-empty host allowlist are required: a flag set with nothing
// allowlisted would enable a fetch path that can reach nothing, which is a
// confusing half-state rather than a safe one. This mirrors
// shouldExposeMetrics' shape (server/graphql.go), which already ANDs a
// flag against a non-empty configuration value for the same reason.
func (s *graphQLServer) providerEnabled() bool {
	return s != nil && s.cfg.DeckProviderEnabled && len(s.deckProviderAllowedHosts) > 0
}

// -----------------------------------------------------------------------
// D-14 checkpoint outcome: Moxfield adapter scaffold (plan 01-07, task 3)
// -----------------------------------------------------------------------
//
// The human selected Moxfield at the task 2 checkpoint
// (docs/research/deck-provider-feasibility.md section 5), diverging from
// task 1's neutral recommendation of Archidekt. What follows is
// deliberately NOT a working Moxfield import: the feasibility record's
// section 2 documents that no Moxfield response body was ever observed --
// api.moxfield.com/robots.txt is a blanket "Disallow: /", and
// moxfield.com's web frontend returned a Cloudflare bot-management 403 on
// every probe attempted across three User-Agents. Fabricating a field
// mapping for a shape that was never seen would be worse than shipping
// nothing: a wrong mapping fails silently by producing a different deck,
// while an explicit unverified-contract seam fails loudly every time it is
// reached.
//
// What IS real below: hostname-keyed routing through this file's secure
// fetch client, gated by the same default-off kill switch providerEnabled
// already enforces, with an explicit seam a future author fills in once an
// authorized sample response exists. No live request to any Moxfield host
// is made by this repository's tests, or by any code path reachable with
// the kill switch at its default (off).

// errMoxfieldContractUnverified is the sentinel normalizeToDeckText
// returns on every call: a compile-time-visible placeholder for "the
// response shape has never been observed," not a bug to be fixed by
// guessing at field names. It must never reach a client -- previewDeckURL
// (server/deck_import.go) maps it onto the same generic paste-fallback
// message every other provider failure already uses (T-01-11), and this
// error's own text carries no provider response content, so logging it
// carries nothing sensitive either.
var errMoxfieldContractUnverified = errors.New("deck provider: moxfield response contract is unverified -- api.moxfield.com/robots.txt disallows automated access and no authorized sample response has ever been captured; see docs/research/deck-provider-feasibility.md section 2")

// moxfieldHost is the exact, lower-cased hostname a pasted Moxfield deck
// URL is expected to name. Deliberately the user-facing web host
// (moxfield.com) rather than the api.moxfield.com data host the
// feasibility record found blocked by robots.txt: this is the host a
// player would actually paste, and the value this file's own
// TestProvider_KillSwitch/TestProvider_KillSwitchRequiresAllowlist tests
// (plan 01-06) already used before this task existed.
const moxfieldHost = "moxfield.com"

// deckProviderAdapter converts one provider's raw fetch response into
// decklist text the canonical parser (pkg/deckimport.Parse) can consume.
// Keyed by hostname in deckProviderAdapters below, so a second provider in
// a later milestone is an added map entry, not a rewrite of
// previewDeckURL (server/deck_import.go).
type deckProviderAdapter interface {
	// source identifies the deckimport.SourceType a successful
	// normalization through this adapter should be attributed to.
	source() deckimport.SourceType
	// normalizeToDeckText converts body into decklist text. An adapter
	// whose response contract has never been observed from a real
	// authorized response MUST return an error naming that fact rather
	// than a fabricated mapping -- see moxfieldAdapter below.
	normalizeToDeckText(body []byte) (string, error)
}

// moxfieldAdapter is deliberately incomplete: its only implemented
// behaviour is refusing to guess. Enabling this provider for a real
// import (not merely routing) requires replacing normalizeToDeckText's
// body with a mapping derived from a real, authorized Moxfield API
// response -- see the two open blockers recorded in
// docs/research/deck-provider-feasibility.md section 5.
type moxfieldAdapter struct{}

func (moxfieldAdapter) source() deckimport.SourceType { return deckimport.SourceMoxfield }

func (moxfieldAdapter) normalizeToDeckText([]byte) (string, error) {
	return "", errMoxfieldContractUnverified
}

// deckProviderAdapters is the hostname-keyed adapter registry
// previewDeckURL consults. A host with no entry here is unsupported
// regardless of whether it appears in the operator's
// DECK_PROVIDER_ALLOWED_HOSTS configuration: the allowlist bounds what
// the secure client may dial, and this map bounds what this codebase
// knows how to interpret once it gets there. These are two independent
// gates, deliberately not merged into one -- the same separation of
// concerns safeControl and newDeckProviderCheckRedirect already apply to
// address policy and host policy respectively.
var deckProviderAdapters = map[string]deckProviderAdapter{
	moxfieldHost: moxfieldAdapter{},
}

// deckProviderAdapterFor returns the adapter registered for rawURL's
// lower-cased hostname, and whether one was found. An unparseable URL or
// an empty host both report "not found" -- previewDeckURL treats either
// the same as an unsupported host, never as a special case needing its
// own message.
func deckProviderAdapterFor(rawURL string) (deckProviderAdapter, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return nil, false
	}
	adapter, ok := deckProviderAdapters[host]
	return adapter, ok
}
