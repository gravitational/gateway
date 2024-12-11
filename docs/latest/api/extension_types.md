# API Reference

## Packages
- [gateway.envoyproxy.io/v1alpha1](#gatewayenvoyproxyiov1alpha1)


## gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the gateway.envoyproxy.io API group.


### Resource Types
- [AuthenticationFilter](#authenticationfilter)
- [RateLimitFilter](#ratelimitfilter)



## AuthenticationFilter









| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1` | | |
| `kind` _string_ | `AuthenticationFilter` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[AuthenticationFilterSpec](#authenticationfilterspec)_ | Spec defines the desired state of the AuthenticationFilter type. |  |  |


## AuthenticationFilterSpec



AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.



_Appears in:_
- [AuthenticationFilter](#authenticationfilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[AuthenticationFilterType](#authenticationfiltertype)_ | Type defines the type of authentication provider to use. Supported provider types<br />are "JWT". |  | Enum: [JWT] <br /> |
| `jwtProviders` _[JwtAuthenticationFilterProvider](#jwtauthenticationfilterprovider) array_ | JWT defines the JSON Web Token (JWT) authentication provider type. When multiple<br />jwtProviders are specified, the JWT is considered valid if any of the providers<br />successfully validate the JWT. For additional details, see<br />https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html. |  | MaxItems: 4 <br /> |


## AuthenticationFilterType

_Underlying type:_ _string_

AuthenticationFilterType is a type of authentication provider.

_Validation:_
- Enum: [JWT]

_Appears in:_
- [AuthenticationFilterSpec](#authenticationfilterspec)

| Field | Description |
| --- | --- |
| `JWT` | JwtAuthenticationFilterProviderType is a provider that uses JSON Web Token (JWT)<br />for authenticating requests..<br /> |


## GlobalRateLimit



GlobalRateLimit defines global rate limit configuration.



_Appears in:_
- [RateLimitFilterSpec](#ratelimitfilterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `rules` _[RateLimitRule](#ratelimitrule) array_ | Rules are a list of RateLimit selectors and limits.<br />Each rule and its associated limit is applied<br />in a mutually exclusive way i.e. if multiple<br />rules get selected, each of their associated<br />limits get applied, so a single traffic request<br />might increase the rate limit counters for multiple<br />rules if selected. |  | MaxItems: 16 <br /> |


## HeaderMatch



HeaderMatch defines the match attributes within the HTTP Headers of the request.



_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[HeaderMatchType](#headermatchtype)_ | Type specifies how to match against the value of the header. | Exact | Enum: [Exact RegularExpression Distinct] <br /> |
| `name` _string_ | Name of the HTTP header. |  | MaxLength: 256 <br />MinLength: 1 <br /> |
| `value` _string_ | Value within the HTTP header. Due to the<br />case-insensitivity of header names, "foo" and "Foo" are considered equivalent.<br />Do not set this field when Type="Distinct", implying matching on any/all unique<br />values within the header. |  | MaxLength: 1024 <br /> |




## JwtAuthenticationFilterProvider



JwtAuthenticationFilterProvider defines the JSON Web Token (JWT) authentication provider type
and how JWTs should be verified:



_Appears in:_
- [AuthenticationFilterSpec](#authenticationfilterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name defines a unique name for the JWT provider. A name can have a variety of forms,<br />including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `issuer` _string_ | Issuer is the principal that issued the JWT and takes the form of a URL or email address.<br />For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.1 for<br />URL format and https://rfc-editor.org/rfc/rfc5322.html for email format. If not provided,<br />the JWT issuer is not checked. |  | MaxLength: 253 <br /> |
| `audiences` _string array_ | Audiences is a list of JWT audiences allowed access. For additional details, see<br />https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences<br />are not checked. |  | MaxItems: 8 <br /> |
| `remoteJWKS` _[RemoteJWKS](#remotejwks)_ | RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote<br />HTTP/HTTPS endpoint. |  |  |


## RateLimitFilter



RateLimitFilter allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1` | | |
| `kind` _string_ | `RateLimitFilter` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[RateLimitFilterSpec](#ratelimitfilterspec)_ | Spec defines the desired state of RateLimitFilter. |  |  |


## RateLimitFilterSpec



RateLimitFilterSpec defines the desired state of RateLimitFilter.



_Appears in:_
- [RateLimitFilter](#ratelimitfilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[RateLimitType](#ratelimittype)_ | Type decides the scope for the RateLimits.<br />Valid RateLimitType values are "Global". |  | Enum: [Global] <br /> |
| `global` _[GlobalRateLimit](#globalratelimit)_ | Global defines global rate limit configuration. |  |  |


## RateLimitRule



RateLimitRule defines the semantics for matching attributes
from the incoming requests, and setting limits for them.



_Appears in:_
- [GlobalRateLimit](#globalratelimit)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientSelectors` _[RateLimitSelectCondition](#ratelimitselectcondition) array_ | ClientSelectors holds the list of select conditions to select<br />specific clients using attributes from the traffic flow.<br />All individual select conditions must hold True for this rule<br />and its limit to be applied.<br />If this field is empty, it is equivalent to True, and<br />the limit is applied. |  | MaxItems: 8 <br /> |
| `limit` _[RateLimitValue](#ratelimitvalue)_ | Limit holds the rate limit values.<br />This limit is applied for traffic flows when the selectors<br />compute to True, causing the request to be counted towards the limit.<br />The limit is enforced and the request is ratelimited, i.e. a response with<br />429 HTTP status code is sent back to the client when<br />the selected requests have reached the limit. |  |  |


## RateLimitSelectCondition



RateLimitSelectCondition specifies the attributes within the traffic flow that can
be used to select a subset of clients to be ratelimited.
All the individual conditions must hold True for the overall condition to hold True.



_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `headers` _[HeaderMatch](#headermatch) array_ | Headers is a list of request headers to match. Multiple header values are ANDed together,<br />meaning, a request MUST match all the specified headers. |  | MaxItems: 16 <br /> |
| `sourceIP` _string_ | SourceIP is the IP CIDR that represents the range of Source IP Addresses of the client.<br />These could also be the intermediate addresses through which the request has flown through and is part of the  `X-Forwarded-For` header.<br />For example, `192.168.0.1/32`, `192.168.0.0/24`, `001:db8::/64`.<br />All IP Addresses within the specified SourceIP CIDR are treated as a single client selector and share the same rate limit bucket. |  |  |


## RateLimitType

_Underlying type:_ _string_

RateLimitType specifies the types of RateLimiting.

_Validation:_
- Enum: [Global]

_Appears in:_
- [RateLimitFilterSpec](#ratelimitfilterspec)

| Field | Description |
| --- | --- |
| `Global` | GlobalRateLimitType allows the rate limits to be applied across all Envoy proxy instances.<br /> |


## RateLimitUnit

_Underlying type:_ _string_

RateLimitUnit specifies the intervals for setting rate limits.
Valid RateLimitUnit values are "Second", "Minute", "Hour", and "Day".

_Validation:_
- Enum: [Second Minute Hour Day]

_Appears in:_
- [RateLimitValue](#ratelimitvalue)



## RateLimitValue



RateLimitValue defines the limits for rate limiting.



_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `requests` _integer_ |  |  |  |
| `unit` _[RateLimitUnit](#ratelimitunit)_ |  |  | Enum: [Second Minute Hour Day] <br /> |


## RemoteJWKS



RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
HTTP/HTTPS endpoint.



_Appears in:_
- [JwtAuthenticationFilterProvider](#jwtauthenticationfilterprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `uri` _string_ | URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to<br />validate the server certificate. |  | MaxLength: 253 <br />MinLength: 1 <br /> |


