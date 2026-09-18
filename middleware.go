package plain

import (
	"net/http"
	"net/netip"
	"strings"
)

const (
	// HeaderXRealIP is the X-Real-IP header.
	HeaderXRealIP = "X-Real-IP"

	// HeaderXForwardedFor is the X-Forwarded-For header.
	HeaderXForwardedFor = "X-Forwarded-For"
)

type realIPConfig struct {
	proxies []netip.Prefix
	headers []string
}

type realIPOption func(*realIPConfig)

func (r *realIPConfig) CanBeTrusted(request *http.Request) bool {
	if len(r.proxies) == 0 {
		return true
	}

	address, _, _, ok := parseRemoteAddress(request.RemoteAddr)
	if !ok {
		return false
	}

	return r.canBeTrusted(address)
}

func (r *realIPConfig) UpdateIP(request *http.Request) {
	address, ok := r.headerAddress(request)
	if !ok {
		return
	}

	_, port, hasPort, _ := parseRemoteAddress(request.RemoteAddr)
	updateRemoteAddress(request, address, port, hasPort)
}

func (r *realIPConfig) apply(request *http.Request) {
	if len(r.proxies) == 0 {
		r.UpdateIP(request)

		return
	}

	remoteAddress, port, hasPort, ok := parseRemoteAddress(request.RemoteAddr)
	if !ok || !r.canBeTrusted(remoteAddress) {
		return
	}

	address, ok := r.headerAddress(request)
	if !ok {
		return
	}

	updateRemoteAddress(request, address, port, hasPort)
}

func (r *realIPConfig) canBeTrusted(address netip.Addr) bool {
	address = address.Unmap()

	for index := range r.proxies {
		proxy := r.proxies[index]

		if proxy.IsValid() && proxy.Contains(address) {
			return true
		}
	}

	return false
}

func (r *realIPConfig) headerAddress(request *http.Request) (netip.Addr, bool) {
	for _, header := range r.headers {
		address, ok := parseIPHeader(request, header)
		if !ok {
			continue
		}

		return address, true
	}

	return netip.Addr{}, false
}

// WithTrustedHeaders sets the headers used to determine the client's IP address.
func WithTrustedHeaders(headers ...string) realIPOption {
	canonicalHeaders := make([]string, len(headers))

	for index, header := range headers {
		canonicalHeaders[index] = http.CanonicalHeaderKey(header)
	}

	return func(ri *realIPConfig) {
		ri.headers = canonicalHeaders
	}
}

// WithTrustedProxies restricts trusted headers to requests from the given IP
// addresses or CIDR ranges. Invalid proxies are ignored.
func WithTrustedProxies(proxies ...string) realIPOption {
	trustedProxies := make([]netip.Prefix, len(proxies))

	for index, proxy := range proxies {
		prefix, err := netip.ParsePrefix(proxy)
		if err == nil {
			trustedProxies[index] = normalizePrefix(prefix)

			continue
		}

		address, err := netip.ParseAddr(proxy)
		if err != nil || address.Zone() != "" {
			continue
		}

		address = address.Unmap()
		trustedProxies[index] = netip.PrefixFrom(address, address.BitLen())
	}

	return func(ri *realIPConfig) {
		ri.proxies = trustedProxies
	}
}

// RealIPMiddleware replaces the request's remote address with the first valid IP
// found in the trusted headers. It defaults to X-Real-IP and X-Forwarded-For
// for requests originating from 127.0.0.1/32.
func RealIPMiddleware(opts ...realIPOption) func(http.Handler) http.Handler {
	cfg := realIPConfig{
		proxies: []netip.Prefix{
			netip.PrefixFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), 32),
		},
		headers: []string{
			HeaderXRealIP,
			HeaderXForwardedFor,
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			cfg.apply(request)

			next.ServeHTTP(writer, request)
		})
	}
}

func parseIPHeader(request *http.Request, header string) (netip.Addr, bool) {
	values := request.Header[header]
	if len(values) == 0 {
		return netip.Addr{}, false
	}

	value := values[0]
	value, _, _ = strings.Cut(value, ",")

	value = strings.TrimSpace(value)
	if value == "" {
		return netip.Addr{}, false
	}

	address, err := netip.ParseAddr(value)
	if err != nil || address.Zone() != "" {
		return netip.Addr{}, false
	}

	return address.Unmap(), true
}

func parseRemoteAddress(remoteAddress string) (netip.Addr, uint16, bool, bool) {
	addressPort, err := netip.ParseAddrPort(remoteAddress)
	if err == nil {
		address := addressPort.Addr()

		return address, addressPort.Port(), true, address.Zone() == ""
	}

	address, err := netip.ParseAddr(remoteAddress)
	if err != nil {
		return netip.Addr{}, 0, false, false
	}

	return address, 0, false, address.Zone() == ""
}

func normalizePrefix(prefix netip.Prefix) netip.Prefix {
	address := prefix.Addr()
	bits := prefix.Bits()

	if address.Is4In6() && bits >= 96 {
		return netip.PrefixFrom(address.Unmap(), bits-96).Masked()
	}

	return prefix.Masked()
}

func updateRemoteAddress(request *http.Request, address netip.Addr, port uint16, hasPort bool) {
	if hasPort {
		request.RemoteAddr = netip.AddrPortFrom(address, port).String()

		return
	}

	request.RemoteAddr = address.String()
}
