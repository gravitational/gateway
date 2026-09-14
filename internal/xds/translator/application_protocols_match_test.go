// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func tcpListener() *listenerv3.Listener {
	return &listenerv3.Listener{
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: 443,
					},
				},
			},
		},
	}
}

func tlsInspectors(l *listenerv3.Listener) []*listenerv3.ListenerFilter {
	var found []*listenerv3.ListenerFilter
	for _, filter := range l.ListenerFilters {
		if filter.Name == wellknown.TlsInspector {
			found = append(found, filter)
		}
	}
	return found
}

func TestAddApplicationProtocolsMatch(t *testing.T) {
	tests := []struct {
		name               string
		xdsListener        *listenerv3.Listener
		existingMatch      *listenerv3.FilterChainMatch
		protocols          []string
		expectFilterChain  bool
		expectTLSInspector bool
		expectAppProtocols []string
		expectServerNames  []string
	}{
		{
			name:        "nil listener",
			xdsListener: nil,
			protocols:   []string{"h2"},
		},
		{
			name:        "no protocols",
			xdsListener: tcpListener(),
			protocols:   nil,
		},
		{
			name:        "empty protocol list",
			xdsListener: tcpListener(),
			protocols:   []string{},
		},
		{
			name:        "wildcard protocol is not matched on",
			xdsListener: tcpListener(),
			protocols:   []string{"*"},
		},
		{
			name:               "creates a filter chain match when there is none",
			xdsListener:        tcpListener(),
			protocols:          []string{"teleport-proxy-ssh"},
			expectFilterChain:  true,
			expectTLSInspector: true,
			expectAppProtocols: []string{"teleport-proxy-ssh"},
		},
		{
			name:               "multiple protocols",
			xdsListener:        tcpListener(),
			protocols:          []string{"h2", "http/1.1"},
			expectFilterChain:  true,
			expectTLSInspector: true,
			expectAppProtocols: []string{"h2", "http/1.1"},
		},
		{
			name:               "keeps server names already on the filter chain match",
			xdsListener:        tcpListener(),
			existingMatch:      &listenerv3.FilterChainMatch{ServerNames: []string{"example.com"}},
			protocols:          []string{"h2"},
			expectFilterChain:  true,
			expectTLSInspector: true,
			expectAppProtocols: []string{"h2"},
			expectServerNames:  []string{"example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterChain := &listenerv3.FilterChain{FilterChainMatch: tt.existingMatch}

			err := addApplicationProtocolsMatch(tt.xdsListener, filterChain, tt.protocols)
			require.NoError(t, err)

			if tt.expectFilterChain {
				require.NotNil(t, filterChain.FilterChainMatch)
				require.Equal(t, tt.expectAppProtocols, filterChain.FilterChainMatch.ApplicationProtocols)
				require.Equal(t, tt.expectServerNames, filterChain.FilterChainMatch.ServerNames)
			} else {
				require.Equal(t, tt.existingMatch, filterChain.FilterChainMatch)
			}

			if tt.xdsListener == nil {
				return
			}
			if tt.expectTLSInspector {
				require.Len(t, tlsInspectors(tt.xdsListener), 1, "TLS inspector filter should be added exactly once")
			} else {
				require.Empty(t, tlsInspectors(tt.xdsListener), "TLS inspector filter should not be added")
			}
		})
	}
}

// A nil filter chain is a programming error rather than a supported input, but the guard
// exists so it degrades to a no-op instead of panicking.
func TestAddApplicationProtocolsMatchNilFilterChain(t *testing.T) {
	l := tcpListener()
	require.NoError(t, addApplicationProtocolsMatch(l, nil, []string{"h2"}))
	require.Empty(t, tlsInspectors(l))
}

// addApplicationProtocolsMatch runs after addServerNamesMatch, so it must add its ALPN match to
// the filter chain match that call built rather than replacing it.
func TestAddApplicationProtocolsMatchAfterServerNames(t *testing.T) {
	l := tcpListener()
	filterChain := &listenerv3.FilterChain{}

	require.NoError(t, addServerNamesMatch(l, filterChain, []string{"example.com"}, nil))
	require.NoError(t, addApplicationProtocolsMatch(l, filterChain, []string{"h2"}))

	require.Equal(t, []string{"example.com"}, filterChain.FilterChainMatch.ServerNames)
	require.Equal(t, []string{"h2"}, filterChain.FilterChainMatch.ApplicationProtocols)
	require.Len(t, tlsInspectors(l), 1, "the two calls should share one TLS inspector filter")
}

// A wildcard SNI leaves addServerNamesMatch with nothing to match on, so it builds no filter
// chain match and adds no TLS inspector. ALPN has to supply both.
func TestAddApplicationProtocolsMatchWildcardServerNames(t *testing.T) {
	l := tcpListener()
	filterChain := &listenerv3.FilterChain{}

	require.NoError(t, addServerNamesMatch(l, filterChain, []string{"*"}, nil))
	require.Nil(t, filterChain.FilterChainMatch)
	require.Empty(t, tlsInspectors(l))

	require.NoError(t, addApplicationProtocolsMatch(l, filterChain, []string{"teleport-proxy-ssh"}))

	require.NotNil(t, filterChain.FilterChainMatch)
	require.Empty(t, filterChain.FilterChainMatch.ServerNames)
	require.Equal(t, []string{"teleport-proxy-ssh"}, filterChain.FilterChainMatch.ApplicationProtocols)
	require.Len(t, tlsInspectors(l), 1)
}

// addXdsTLSInspectorFilter returns early when the filter is already present, so passing nil
// fingerprints here must leave the fingerprinting addServerNamesMatch configured intact.
func TestAddApplicationProtocolsMatchPreservesFingerprints(t *testing.T) {
	l := tcpListener()
	filterChain := &listenerv3.FilterChain{}

	fingerprints := []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA3, ir.TLSFingerprintTypeJA4}
	require.NoError(t, addServerNamesMatch(l, filterChain, []string{"example.com"}, fingerprints))
	require.NoError(t, addApplicationProtocolsMatch(l, filterChain, []string{"h2"}))

	filters := tlsInspectors(l)
	require.Len(t, filters, 1)

	inspector := &tls_inspectorv3.TlsInspector{}
	require.NoError(t, filters[0].GetTypedConfig().UnmarshalTo(inspector))
	require.True(t, inspector.GetEnableJa3Fingerprinting().GetValue(), "JA3 fingerprinting should survive")
	require.True(t, inspector.GetEnableJa4Fingerprinting().GetValue(), "JA4 fingerprinting should survive")
}
