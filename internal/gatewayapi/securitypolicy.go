// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"

	perr "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	defaultRedirectURL        = "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback"
	defaultRedirectPath       = "/oauth2/callback"
	defaultLogoutPath         = "/logout"
	defaultForwardAccessToken = false
	defaultRefreshToken       = false

	// nolint: gosec
	oidcHMACSecretName = "envoy-oidc-hmac"
	oidcHMACSecretKey  = "hmac-secret"
)

func (t *Translator) ProcessSecurityPolicies(securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*egv1a1.SecurityPolicy {
	var res []*egv1a1.SecurityPolicy

	// Sort based on timestamp
	sort.Slice(securityPolicies, func(i, j int) bool {
		return securityPolicies[i].CreationTimestamp.Before(&(securityPolicies[j].CreationTimestamp))
	})

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMap := map[policyTargetRouteKey]*policyRouteTargetContext{}
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(GetRouteType(route)),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Map of Gateway to the routes attached to it
	gatewayRouteMap := make(map[string]sets.Set[string])

	handledPolicies := make(map[types.NamespacedName]*egv1a1.SecurityPolicy)

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway {
				var (
					targetedRoute  RouteContext
					parentGateways []gwapiv1a2.ParentReference
					resolveErr     *status.PolicyResolveError
				)
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				targetedRoute, resolveErr = resolveSecurityPolicyRouteTargetRef(policy, currTarget, routeMap)
				// Skip if the route is not found
				// It's not necessarily an error because the SecurityPolicy may be
				// reconciled by multiple controllers. And the other controller may
				// have the target route.
				if targetedRoute == nil {
					continue
				}

				// Find the parent Gateways for the route and add it to the
				// gatewayRouteMap, which will be used to check policy override.
				// The parent gateways are also used to set the status of the policy.
				parentRefs := GetParentReferences(targetedRoute)
				for _, p := range parentRefs {
					if p.Kind == nil || *p.Kind == resource.KindGateway {
						namespace := targetedRoute.GetNamespace()
						if p.Namespace != nil {
							namespace = string(*p.Namespace)
						}
						gwNN := types.NamespacedName{
							Namespace: namespace,
							Name:      string(p.Name),
						}

						key := gwNN.String()
						if _, ok := gatewayRouteMap[key]; !ok {
							gatewayRouteMap[key] = make(sets.Set[string])
						}
						gatewayRouteMap[key].Insert(utils.NamespacedName(targetedRoute).String())
						parentGateways = append(parentGateways, getAncestorRefForPolicy(gwNN, p.SectionName))
					}
				}

				// Set conditions for resolve error, then skip current xroute
				if resolveErr != nil {
					status.SetResolveErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						resolveErr,
					)

					continue
				}

				if err := t.translateSecurityPolicyForRoute(policy, targetedRoute, resources, xdsIR); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(err),
					)
				}

				// Set Accepted condition if it is unset
				status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName)
			}
		}
	}

	// Process the policies targeting Gateways
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway {
				var (
					targetedGateway *GatewayContext
					resolveErr      *status.PolicyResolveError
				)

				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				targetedGateway, resolveErr = resolveSecurityPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
				// Skip if the gateway is not found
				// It's not necessarily an error because the SecurityPolicy may be
				// reconciled by multiple controllers. And the other controller may
				// have the target gateway.
				if targetedGateway == nil {
					continue
				}

				// Find its ancestor reference by resolved gateway, even with resolve error
				gatewayNN := utils.NamespacedName(targetedGateway)
				parentGateways := []gwapiv1a2.ParentReference{
					getAncestorRefForPolicy(gatewayNN, nil),
				}

				// Set conditions for resolve error, then skip current gateway
				if resolveErr != nil {
					status.SetResolveErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						resolveErr,
					)

					continue
				}

				if err := t.translateSecurityPolicyForGateway(policy, targetedGateway, currTarget, resources, xdsIR); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(err),
					)
				}

				// Set Accepted condition if it is unset
				status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName)

				// Check if this policy is overridden by other policies targeting
				// at route level
				if r, ok := gatewayRouteMap[gatewayNN.String()]; ok {
					// Maintain order here to ensure status/string does not change with the same data
					routes := r.UnsortedList()
					sort.Strings(routes)
					message := fmt.Sprintf(
						"This policy is being overridden by other securityPolicies for these routes: %v",
						routes)
					status.SetConditionForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						egv1a1.PolicyConditionOverridden,
						metav1.ConditionTrue,
						egv1a1.PolicyReasonOverridden,
						message,
						policy.Generation,
					)
				}
			}
		}
	}

	return res
}

func resolveSecurityPolicyGatewayTargetRef(
	policy *egv1a1.SecurityPolicy,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	gateways map[types.NamespacedName]*policyGatewayTargetContext,
) (*GatewayContext, *status.PolicyResolveError) {
	// Find the Gateway
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}
	gateway, ok := gateways[key]

	// Gateway not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := fmt.Sprintf("Unable to target Gateway %s, another SecurityPolicy has already attached to it",
			string(target.Name))

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	gateway.attached = true
	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveSecurityPolicyRouteTargetRef(
	policy *egv1a1.SecurityPolicy,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	routes map[policyTargetRouteKey]*policyRouteTargetContext,
) (RouteContext, *status.PolicyResolveError) {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}
	route, ok := routes[key]

	// Route not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s %s, another SecurityPolicy has already attached to it",
			string(target.Kind), string(target.Name))

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	route.attached = true
	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateSecurityPolicyForRoute(
	policy *egv1a1.SecurityPolicy, route RouteContext,
	resources *resource.Resources, xdsIR resource.XdsIRMap,
) error {
	// Build IR
	var (
		cors          *ir.CORS
		jwt           *ir.JWT
		basicAuth     *ir.BasicAuth
		authorization *ir.Authorization
		err, errs     error
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.JWT != nil {
		jwt = t.buildJWT(policy.Spec.JWT)
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			resources); err != nil {
			err = perr.WithMessage(err, "BasicAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	parentRefs := GetParentReferences(route)
	for _, p := range parentRefs {
		parentRefCtx := GetRouteParentContext(route, p)
		gtwCtx := parentRefCtx.GetGateway()
		if gtwCtx == nil {
			continue
		}

		var extAuth *ir.ExtAuth
		if policy.Spec.ExtAuth != nil {
			if extAuth, err = t.buildExtAuth(
				policy,
				resources,
				gtwCtx.envoyProxy); err != nil {
				err = perr.WithMessage(err, "ExtAuth")
				errs = errors.Join(errs, err)
			}
		}

		var oidc *ir.OIDC
		if policy.Spec.OIDC != nil {
			if oidc, err = t.buildOIDC(
				policy,
				resources,
				gtwCtx.envoyProxy); err != nil { // TODO zhaohuabing: Only the last EnvoyProxy is used
				err = perr.WithMessage(err, "OIDC")
				errs = errors.Join(errs, err)
			}
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		for _, listener := range parentRefCtx.listeners {
			irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
			if irListener != nil {
				for _, r := range irListener.Routes {
					if strings.HasPrefix(r.Name, prefix) {
						r.Security = &ir.SecurityFeatures{
							CORS:          cors,
							JWT:           jwt,
							OIDC:          oidc,
							BasicAuth:     basicAuth,
							ExtAuth:       extAuth,
							Authorization: authorization,
						}
						if errs != nil {
							// Return a 500 direct response to avoid unauthorized access
							r.DirectResponse = &ir.CustomResponse{
								StatusCode: ptr.To(uint32(500)),
							}
						}
					}
				}
			}
		}
	}
	return errs
}

func (t *Translator) translateSecurityPolicyForGateway(
	policy *egv1a1.SecurityPolicy,
	gateway *GatewayContext,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) error {
	// Build IR
	var (
		cors          *ir.CORS
		jwt           *ir.JWT
		oidc          *ir.OIDC
		basicAuth     *ir.BasicAuth
		extAuth       *ir.ExtAuth
		authorization *ir.Authorization
		err, errs     error
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.JWT != nil {
		jwt = t.buildJWT(policy.Spec.JWT)
	}

	if policy.Spec.OIDC != nil {
		if oidc, err = t.buildOIDC(
			policy,
			resources,
			gateway.envoyProxy); err != nil {
			err = perr.WithMessage(err, "OIDC")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			resources); err != nil {
			err = perr.WithMessage(err, "BasicAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.ExtAuth != nil {
		if extAuth, err = t.buildExtAuth(
			policy,
			resources,
			gateway.envoyProxy); err != nil {
			err = perr.WithMessage(err, "ExtAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	// Apply IR to all the routes within the specific Gateway that originated
	// from the gateway to which this security policy was attached.
	// If the feature is already set, then skip it, since it must have be
	// set by a policy attaching to the route
	//
	// Note: there are multiple features in a security policy, even if some of them
	// are invalid, we still want to apply the valid ones.
	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(target.Name))
	for _, h := range x.HTTP {
		gatewayName := h.Name[0:strings.LastIndex(h.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}
		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range h.Routes {
			// If any of the features are already set, it means that a more specific
			// policy(targeting xRoute) has already set it, so we skip it.
			if r.Security != nil {
				continue
			}
			r.Security = &ir.SecurityFeatures{
				CORS:          cors,
				JWT:           jwt,
				OIDC:          oidc,
				BasicAuth:     basicAuth,
				ExtAuth:       extAuth,
				Authorization: authorization,
			}
			if errs != nil {
				// Return a 500 direct response to avoid unauthorized access
				r.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			}
		}
	}
	return errs
}

func (t *Translator) buildCORS(cors *egv1a1.CORS) *ir.CORS {
	var allowOrigins []*ir.StringMatch

	for _, origin := range cors.AllowOrigins {
		if isWildcard(string(origin)) {
			regexStr := wildcard2regex(string(origin))
			allowOrigins = append(allowOrigins, &ir.StringMatch{
				SafeRegex: &regexStr,
			})
		} else {
			allowOrigins = append(allowOrigins, &ir.StringMatch{
				Exact: (*string)(&origin),
			})
		}
	}

	return &ir.CORS{
		AllowOrigins:     allowOrigins,
		AllowMethods:     cors.AllowMethods,
		AllowHeaders:     cors.AllowHeaders,
		ExposeHeaders:    cors.ExposeHeaders,
		MaxAge:           cors.MaxAge,
		AllowCredentials: cors.AllowCredentials != nil && *cors.AllowCredentials,
	}
}

func isWildcard(s string) bool {
	return strings.ContainsAny(s, "*")
}

func wildcard2regex(wildcard string) string {
	regexStr := strings.ReplaceAll(wildcard, ".", "\\.")
	regexStr = strings.ReplaceAll(regexStr, "*", ".*")
	return regexStr
}

func (t *Translator) buildJWT(jwt *egv1a1.JWT) *ir.JWT {
	return &ir.JWT{
		AllowMissing: ptr.Deref(jwt.Optional, false),
		Providers:    jwt.Providers,
	}
}

func (t *Translator) buildOIDC(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.OIDC, error) {
	var (
		oidc               = policy.Spec.OIDC
		provider           *ir.OIDCProvider
		clientSecret       *corev1.Secret
		redirectURL        = defaultRedirectURL
		redirectPath       = defaultRedirectPath
		logoutPath         = defaultLogoutPath
		forwardAccessToken = defaultForwardAccessToken
		refreshToken       = defaultRefreshToken
		err                error
	)

	if provider, err = t.buildOIDCProvider(policy, resources, envoyProxy); err != nil {
		return nil, err
	}

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: policy.Namespace,
	}

	if clientSecret, err = t.validateSecretRef(
		false, from, oidc.ClientSecret, resources); err != nil {
		return nil, err
	}

	clientSecretBytes, ok := clientSecret.Data[egv1a1.OIDCClientSecretKey]
	if !ok || len(clientSecretBytes) == 0 {
		return nil, fmt.Errorf(
			"client secret not found in secret %s/%s",
			clientSecret.Namespace, clientSecret.Name)
	}

	scopes := appendOpenidScopeIfNotExist(oidc.Scopes)

	if oidc.RedirectURL != nil {
		path, err := extractRedirectPath(*oidc.RedirectURL)
		if err != nil {
			return nil, err
		}
		redirectURL = *oidc.RedirectURL
		redirectPath = path
	}
	if oidc.LogoutPath != nil {
		logoutPath = *oidc.LogoutPath
	}
	if oidc.ForwardAccessToken != nil {
		forwardAccessToken = *oidc.ForwardAccessToken
	}
	if oidc.RefreshToken != nil {
		refreshToken = *oidc.RefreshToken
	}

	// Generate a unique cookie suffix for oauth filters.
	// This is to avoid cookie name collision when multiple security policies are applied
	// to the same route.
	suffix := utils.Digest32(string(policy.UID))

	// Get the HMAC secret.
	// HMAC secret is generated by the CertGen job and stored in a secret
	// We need to rotate the HMAC secret in the future, probably the same
	// way we rotate the certs generated by the CertGen job.
	hmacSecret := resources.GetSecret(t.Namespace, oidcHMACSecretName)
	if hmacSecret == nil {
		return nil, fmt.Errorf("HMAC secret %s/%s not found", t.Namespace, oidcHMACSecretName)
	}
	hmacData, ok := hmacSecret.Data[oidcHMACSecretKey]
	if !ok || len(hmacData) == 0 {
		return nil, fmt.Errorf(
			"HMAC secret not found in secret %s/%s", t.Namespace, oidcHMACSecretName)
	}

	return &ir.OIDC{
		Name:                   irConfigName(policy),
		Provider:               *provider,
		ClientID:               oidc.ClientID,
		ClientSecret:           clientSecretBytes,
		Scopes:                 scopes,
		Resources:              oidc.Resources,
		RedirectURL:            redirectURL,
		RedirectPath:           redirectPath,
		LogoutPath:             logoutPath,
		ForwardAccessToken:     forwardAccessToken,
		DefaultTokenTTL:        oidc.DefaultTokenTTL,
		RefreshToken:           refreshToken,
		DefaultRefreshTokenTTL: oidc.DefaultRefreshTokenTTL,
		CookieSuffix:           suffix,
		CookieNameOverrides:    policy.Spec.OIDC.CookieNames,
		CookieDomain:           policy.Spec.OIDC.CookieDomain,
		HMACSecret:             hmacData,
	}, nil
}

func (t *Translator) buildOIDCProvider(policy *egv1a1.SecurityPolicy, resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy) (*ir.OIDCProvider, error) {
	var (
		provider              = policy.Spec.OIDC.Provider
		tokenEndpoint         string
		authorizationEndpoint string
		protocol              ir.AppProtocol
		rd                    *ir.RouteDestination
		traffic               *ir.TrafficFeatures
		err                   error
	)

	// Discover the token and authorization endpoints from the issuer's
	// well-known url if not explicitly specified
	if provider.TokenEndpoint == nil || provider.AuthorizationEndpoint == nil {
		tokenEndpoint, authorizationEndpoint, err = fetchEndpointsFromIssuer(provider.Issuer)
		if err != nil {
			return nil, fmt.Errorf("error fetching endpoints from issuer: %w", err)
		}
	} else {
		tokenEndpoint = *provider.TokenEndpoint
		authorizationEndpoint = *provider.AuthorizationEndpoint
	}

	if err = validateTokenEndpoint(tokenEndpoint); err != nil {
		return nil, err
	}

	u, err := url.Parse(tokenEndpoint)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		protocol = ir.HTTPS
	} else {
		protocol = ir.HTTP
	}

	if len(provider.BackendRefs) > 0 {
		if rd, err = t.translateExtServiceBackendRefs(policy, provider.BackendRefs, protocol, resources, envoyProxy, "oidc", 0); err != nil {
			return nil, err
		}
	}

	if traffic, err = translateTrafficFeatures(provider.BackendSettings); err != nil {
		return nil, err
	}

	return &ir.OIDCProvider{
		Destination:           rd,
		Traffic:               traffic,
		AuthorizationEndpoint: authorizationEndpoint,
		TokenEndpoint:         tokenEndpoint,
	}, nil
}

func extractRedirectPath(redirectURL string) (string, error) {
	schemeDelimiter := strings.Index(redirectURL, "://")
	if schemeDelimiter <= 0 {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	scheme := redirectURL[:schemeDelimiter]
	if scheme != "http" && scheme != "https" && scheme != "%REQ(x-forwarded-proto)%" {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	hostDelimiter := strings.Index(redirectURL[schemeDelimiter+3:], "/")
	if hostDelimiter <= 0 {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	path := redirectURL[schemeDelimiter+3+hostDelimiter:]
	if path == "/" {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	return path, nil
}

// appendOpenidScopeIfNotExist appends the openid scope to the provided scopes
// if it is not already present.
// `openid` is a required scope for OIDC.
// see https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
func appendOpenidScopeIfNotExist(scopes []string) []string {
	const authScopeOpenID = "openid"

	hasOpenIDScope := false
	for _, scope := range scopes {
		if scope == authScopeOpenID {
			hasOpenIDScope = true
		}
	}
	if !hasOpenIDScope {
		scopes = append(scopes, authScopeOpenID)
	}
	return scopes
}

type OpenIDConfig struct {
	TokenEndpoint         string `json:"token_endpoint"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

func fetchEndpointsFromIssuer(issuerURL string) (string, string, error) {
	// Fetch the OpenID configuration from the issuer URL
	resp, err := http.Get(fmt.Sprintf("%s/.well-known/openid-configuration", issuerURL))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Parse the OpenID configuration response
	var config OpenIDConfig
	err = json.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return "", "", err
	}

	return config.TokenEndpoint, config.AuthorizationEndpoint, nil
}

// validateTokenEndpoint validates the token endpoint URL
func validateTokenEndpoint(tokenEndpoint string) error {
	parsedURL, err := url.Parse(tokenEndpoint)
	if err != nil {
		return fmt.Errorf("error parsing token endpoint URL: %w", err)
	}

	if ip, err := netip.ParseAddr(parsedURL.Hostname()); err == nil {
		if ip.Unmap().Is4() {
			return fmt.Errorf("token endpoint URL must be a domain name: %s", tokenEndpoint)
		}
	}

	if parsedURL.Port() != "" {
		_, err = strconv.Atoi(parsedURL.Port())
		if err != nil {
			return fmt.Errorf("error parsing token endpoint URL port: %w", err)
		}
	}
	return nil
}

func (t *Translator) buildBasicAuth(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
) (*ir.BasicAuth, error) {
	var (
		basicAuth   = policy.Spec.BasicAuth
		usersSecret *corev1.Secret
		err         error
	)

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: policy.Namespace,
	}
	if usersSecret, err = t.validateSecretRef(
		false, from, basicAuth.Users, resources); err != nil {
		return nil, err
	}

	usersSecretBytes, ok := usersSecret.Data[egv1a1.BasicAuthUsersSecretKey]
	if !ok || len(usersSecretBytes) == 0 {
		return nil, fmt.Errorf(
			"users secret not found in secret %s/%s",
			usersSecret.Namespace, usersSecret.Name)
	}

	return &ir.BasicAuth{
		Name:  irConfigName(policy),
		Users: usersSecretBytes,
	}, nil
}

func (t *Translator) buildExtAuth(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.ExtAuth, error) {
	var (
		http            = policy.Spec.ExtAuth.HTTP
		grpc            = policy.Spec.ExtAuth.GRPC
		backendRefs     []egv1a1.BackendRef
		backendSettings *egv1a1.ClusterSettings
		protocol        ir.AppProtocol
		rd              *ir.RouteDestination
		authority       string
		err             error
		traffic         *ir.TrafficFeatures
	)

	// These are sanity checks, they should never happen because the API server
	// should have caught them
	if http == nil && grpc == nil {
		return nil, errors.New("one of grpc or http must be specified")
	} else if http != nil && grpc != nil {
		return nil, errors.New("only one of grpc or http can be specified")
	}

	switch {
	case http != nil:
		protocol = ir.HTTP
		switch {
		case len(http.BackendRefs) > 0:
			backendRefs = http.BackendCluster.BackendRefs
		case http.BackendRef != nil:
			backendRefs = []egv1a1.BackendRef{
				{
					BackendObjectReference: *http.BackendRef,
				},
			}
		default:
			// This is a sanity check, it should never happen because the API server should have caught it
			return nil, errors.New("http backend refs must be specified")
		}
	case grpc != nil:
		protocol = ir.GRPC
		switch {
		case len(grpc.BackendCluster.BackendRefs) > 0:
			backendRefs = grpc.BackendRefs
		case grpc.BackendRef != nil:
			backendRefs = []egv1a1.BackendRef{
				{
					BackendObjectReference: *grpc.BackendRef,
				},
			}
		default:
			// This is a sanity check, it should never happen because the API server should have caught it
			return nil, errors.New("grpc backend refs must be specified")
		}
	}

	if rd, err = t.translateExtServiceBackendRefs(policy, backendRefs, protocol, resources, envoyProxy, "extauth", 0); err != nil {
		return nil, err
	}

	for _, backendRef := range backendRefs {
		// Authority is the calculated hostname that will be used as the Authority header.
		// If there are multiple backend referenced, simply use the first one - there are no good answers here.
		// When translated to XDS, the authority is used on the filter level not on the cluster level.
		// There's no way to translate to XDS and use a different authority for each backendref
		if authority == "" {
			authority = backendRefAuthority(resources, &backendRef.BackendObjectReference, policy)
		}
	}

	if traffic, err = translateTrafficFeatures(backendSettings); err != nil {
		return nil, err
	}
	extAuth := &ir.ExtAuth{
		Name:             irConfigName(policy),
		HeadersToExtAuth: policy.Spec.ExtAuth.HeadersToExtAuth,
		FailOpen:         policy.Spec.ExtAuth.FailOpen,
		Traffic:          traffic,
		RecomputeRoute:   policy.Spec.ExtAuth.RecomputeRoute,
	}

	if http != nil {
		extAuth.HTTP = &ir.HTTPExtAuthService{
			Destination:      *rd,
			Authority:        authority,
			Path:             ptr.Deref(http.Path, ""),
			HeadersToBackend: http.HeadersToBackend,
		}
	} else {
		extAuth.GRPC = &ir.GRPCExtAuthService{
			Destination: *rd,
			Authority:   authority,
		}
	}
	return extAuth, nil
}

func backendRefAuthority(resources *resource.Resources, backendRef *gwapiv1.BackendObjectReference, policy *egv1a1.SecurityPolicy) string {
	if backendRef == nil {
		return ""
	}

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, policy.Namespace)
	backendKind := KindDerefOr(backendRef.Kind, resource.KindService)
	if backendKind == resource.KindBackend {
		backend := resources.GetBackend(backendNamespace, string(backendRef.Name))
		if backend != nil {
			// TODO: exists multi FQDN endpoints?
			for _, ep := range backend.Spec.Endpoints {
				if ep.FQDN != nil {
					return net.JoinHostPort(ep.FQDN.Hostname, strconv.Itoa(int(ep.FQDN.Port)))
				}
			}
		}
	}

	return net.JoinHostPort(
		fmt.Sprintf("%s.%s", backendRef.Name, backendNamespace),
		strconv.Itoa(int(*backendRef.Port)),
	)
}

func (t *Translator) buildAuthorization(policy *egv1a1.SecurityPolicy) (*ir.Authorization, error) {
	var (
		authorization = policy.Spec.Authorization
		irAuth        = &ir.Authorization{}
		// The default action is Deny if not specified
		defaultAction = egv1a1.AuthorizationActionDeny
	)

	if authorization.DefaultAction != nil {
		defaultAction = *authorization.DefaultAction
	}
	irAuth.DefaultAction = defaultAction

	for i, rule := range authorization.Rules {
		principal := ir.Principal{}

		for _, cidr := range rule.Principal.ClientCIDRs {
			cidrMatch, err := parseCIDR(string(cidr))
			if err != nil {
				return nil, fmt.Errorf("unable to translate authorization rule: %w", err)
			}

			principal.ClientCIDRs = append(principal.ClientCIDRs, cidrMatch)
		}

		principal.JWT = rule.Principal.JWT

		var name string
		if rule.Name != nil && *rule.Name != "" {
			name = *rule.Name
		} else {
			name = defaultAuthorizationRuleName(policy, i)
		}
		irAuth.Rules = append(irAuth.Rules, &ir.AuthorizationRule{
			Name:      name,
			Action:    rule.Action,
			Principal: principal,
		})
	}

	return irAuth, nil
}

func defaultAuthorizationRuleName(policy *egv1a1.SecurityPolicy, index int) string {
	return fmt.Sprintf(
		"%s/authorization/rule/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}
