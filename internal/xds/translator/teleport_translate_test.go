package translator

import (
	"testing"

	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
)

func TestTeleportTranslateXds(t *testing.T) {
	testCases := []struct {
		name           string
		requireSecrets bool
	}{
		{
			name: "empty",
		},
		{
			name:           "simple-tls",
			requireSecrets: true,
		},
		{
			name: "tls-route-passthrough",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ir := requireXdsIRFromInputTestData(t, "xds-ir", tc.name+".yaml")
			tr := &Translator{
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: infra.GetRateLimitServiceURL("envoy-gateway-system"),
				},
			}

			tCtx, err := tr.Translate(ir)
			require.NoError(t, err)
			listeners := tCtx.XdsResources[resourcev3.ListenerType]
			routes := tCtx.XdsResources[resourcev3.RouteType]
			clusters := tCtx.XdsResources[resourcev3.ClusterType]
			endpoints := tCtx.XdsResources[resourcev3.EndpointType]
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".listeners.yaml"), requireResourcesToYAMLString(t, listeners))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".routes.yaml"), requireResourcesToYAMLString(t, routes))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".clusters.yaml"), requireResourcesToYAMLString(t, clusters))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".endpoints.yaml"), requireResourcesToYAMLString(t, endpoints))
			if tc.requireSecrets {
				secrets := tCtx.XdsResources[resourcev3.SecretType]
				require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}
		})
	}
}
