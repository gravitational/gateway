# API Reference

## Packages
- [config.gateway.envoyproxy.io/v1alpha1](#configgatewayenvoyproxyiov1alpha1)


## config.gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the config.gateway.envoyproxy.io
API group.


### Resource Types
- [EnvoyGateway](#envoygateway)
- [EnvoyProxy](#envoyproxy)



## EnvoyGateway



EnvoyGateway is the schema for the envoygateways API.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `config.gateway.envoyproxy.io/v1alpha1` | | |
| `kind` _string_ | `EnvoyGateway` | | |
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset,<br />default configuration parameters will apply. |  |  |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration.<br />If unspecified, the Kubernetes provider is used with default configuration<br />parameters. |  |  |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service<br />deployed by Envoy Gateway required to implement the Global Rate limiting<br />functionality. The specific rate limit service used here is the reference<br />implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit.<br />This configuration is unneeded for "Local" rate limiting. |  |  |
| `extension` _[Extension](#extension)_ | Extension defines an extension to register for the Envoy Gateway Control Plane. |  |  |


## EnvoyGatewayFileProvider



EnvoyGatewayFileProvider defines configuration for the File provider.



_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)



## EnvoyGatewayKubernetesProvider



EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.



_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `rateLimitDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource.<br />If unspecified, default settings for the manged Envoy ratelimit deployment resource<br />are applied. |  |  |


## EnvoyGatewayProvider



EnvoyGatewayProvider defines the desired configuration of a provider.



_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of provider to use. Supported types are "Kubernetes". |  | Enum: [Kubernetes] <br /> |
| `kubernetes` _[EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)_ | Kubernetes defines the configuration of the Kubernetes provider. Kubernetes<br />provides runtime configuration via the Kubernetes API. |  |  |
| `file` _[EnvoyGatewayFileProvider](#envoygatewayfileprovider)_ | File defines the configuration of the File provider. File provides runtime<br />configuration defined by one or more files. This type is not implemented<br />until https://github.com/envoyproxy/gateway/issues/1001 is fixed. |  |  |


## EnvoyGatewaySpec



EnvoyGatewaySpec defines the desired state of Envoy Gateway.



_Appears in:_
- [EnvoyGateway](#envoygateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset,<br />default configuration parameters will apply. |  |  |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration.<br />If unspecified, the Kubernetes provider is used with default configuration<br />parameters. |  |  |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service<br />deployed by Envoy Gateway required to implement the Global Rate limiting<br />functionality. The specific rate limit service used here is the reference<br />implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit.<br />This configuration is unneeded for "Local" rate limiting. |  |  |
| `extension` _[Extension](#extension)_ | Extension defines an extension to register for the Envoy Gateway Control Plane. |  |  |


## EnvoyProxy



EnvoyProxy is the schema for the envoyproxies API.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `config.gateway.envoyproxy.io/v1alpha1` | | |
| `kind` _string_ | `EnvoyProxy` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[EnvoyProxySpec](#envoyproxyspec)_ | EnvoyProxySpec defines the desired state of EnvoyProxy. |  |  |


## EnvoyProxyKubernetesProvider



EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource
provider.



_Appears in:_
- [EnvoyProxyProvider](#envoyproxyprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `envoyDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | EnvoyDeployment defines the desired state of the Envoy deployment resource.<br />If unspecified, default settings for the manged Envoy deployment resource<br />are applied. |  |  |
| `envoyService` _[KubernetesServiceSpec](#kubernetesservicespec)_ | EnvoyService defines the desired state of the Envoy service resource.<br />If unspecified, default settings for the manged Envoy service resource<br />are applied. |  |  |


## EnvoyProxyProvider



EnvoyProxyProvider defines the desired state of a resource provider.



_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of resource provider to use. A resource provider provides<br />infrastructure resources for running the data plane, e.g. Envoy proxy, and<br />optional auxiliary control planes. Supported types are "Kubernetes". |  | Enum: [Kubernetes] <br /> |
| `kubernetes` _[EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)_ | Kubernetes defines the desired state of the Kubernetes resource provider.<br />Kubernetes provides infrastructure resources for running the data plane,<br />e.g. Envoy proxy. If unspecified and type is "Kubernetes", default settings<br />for managed Kubernetes resources are applied. |  |  |


## EnvoyProxySpec



EnvoyProxySpec defines the desired state of EnvoyProxy.



_Appears in:_
- [EnvoyProxy](#envoyproxy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `provider` _[EnvoyProxyProvider](#envoyproxyprovider)_ | Provider defines the desired resource provider and provider-specific configuration.<br />If unspecified, the "Kubernetes" resource provider is used with default configuration<br />parameters. |  |  |
| `logging` _[ProxyLogging](#proxylogging)_ | Logging defines logging parameters for managed proxies. If unspecified,<br />default settings apply. This type is not implemented until<br />https://github.com/envoyproxy/gateway/issues/280 is fixed. | \{ level:map[system:info] \} |  |
| `bootstrap` _string_ | Bootstrap defines the Envoy Bootstrap as a YAML string.<br />Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap<br />to learn more about the syntax.<br />If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration<br />set by Envoy Gateway.<br />Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources<br />from it are not configurable and will result in the `EnvoyProxy` resource being rejected.<br />Backward compatibility across minor versions is not guaranteed.<br />We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default<br />Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors. |  |  |




## Extension



Extension defines the configuration for registering an extension to
the Envoy Gateway control plane.



_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resources` _[GroupVersionKind](#groupversionkind) array_ | Resources defines the set of K8s resources the extension will handle. |  |  |
| `hooks` _[ExtensionHooks](#extensionhooks)_ | Hooks defines the set of hooks the extension supports |  | Required: \{\} <br /> |
| `service` _[ExtensionService](#extensionservice)_ | Service defines the configuration of the extension service that the Envoy<br />Gateway Control Plane will call through extension hooks. |  | Required: \{\} <br /> |


## ExtensionHooks



ExtensionHooks defines extension hooks across all supported runners



_Appears in:_
- [Extension](#extension)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `xdsTranslator` _[XDSTranslatorHooks](#xdstranslatorhooks)_ | XDSTranslator defines all the supported extension hooks for the xds-translator runner |  |  |


## ExtensionService



ExtensionService defines the configuration for connecting to a registered extension service.



_Appears in:_
- [Extension](#extension)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host define the extension service hostname. |  |  |
| `port` _integer_ | Port defines the port the extension service is exposed on. | 80 | Minimum: 0 <br /> |
| `tls` _[ExtensionTLS](#extensiontls)_ | TLS defines TLS configuration for communication between Envoy Gateway and<br />the extension service. |  |  |


## ExtensionTLS



ExtensionTLS defines the TLS configuration when connecting to an extension service



_Appears in:_
- [ExtensionService](#extensionservice)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `certificateRef` _[SecretObjectReference](#secretobjectreference)_ | CertificateRef contains a references to objects (Kubernetes objects or otherwise) that<br />contains a TLS certificate and private keys. These certificates are used to<br />establish a TLS handshake to the extension server.<br /><br />CertificateRef can only reference a Kubernetes Secret at this time. |  | Required: \{\} <br /> |


## Gateway



Gateway defines the desired Gateway API configuration of Envoy Gateway.



_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `controllerName` _string_ | ControllerName defines the name of the Gateway API controller. If unspecified,<br />defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following<br />for additional details:<br />  https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass |  |  |


## GroupVersionKind



GroupVersionKind unambiguously identifies a Kind.
It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind



_Appears in:_
- [Extension](#extension)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _string_ |  |  |  |
| `version` _string_ |  |  |  |
| `kind` _string_ |  |  |  |


## KubernetesContainerSpec



KubernetesContainerSpec defines the desired state of the Kubernetes container resource.



_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)_ | Resources required by this container.<br />More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |  |  |
| `securityContext` _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#securitycontext-v1-core)_ | SecurityContext defines the security options the container should be run with.<br />If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.<br />More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |  |  |
| `image` _string_ | Image specifies the EnvoyProxy container image to be used, instead of the default image. |  |  |


## KubernetesDeploymentSpec



KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.



_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `patch` _[KubernetesPatchSpec](#kubernetespatchspec)_ | Patch defines how to perform the patch operation to deployment |  |  |
| `replicas` _integer_ | Replicas is the number of desired pods. Defaults to 1. |  |  |
| `pod` _[KubernetesPodSpec](#kubernetespodspec)_ | Pod defines the desired annotations and securityContext of container. |  |  |
| `container` _[KubernetesContainerSpec](#kubernetescontainerspec)_ | Container defines the resources and securityContext of container. |  |  |


## KubernetesPatchSpec



KubernetesPatchSpec defines how to perform the patch operation



_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[MergeType](#mergetype)_ | Type is the type of merge operation to perform<br /><br />By default, StrategicMerge is used as the patch type. |  |  |
| `value` _[JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#json-v1-apiextensions-k8s-io)_ | Object contains the raw configuration for merged object |  |  |


## KubernetesPodSpec



KubernetesPodSpec defines the desired state of the Kubernetes pod resource.



_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations are the annotations that should be appended to the pods.<br />By default, no pod annotations are appended. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings.<br />Optional: Defaults to empty.  See type description for default values of each field. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#affinity-v1-core)_ | If specified, the pod's scheduling constraints. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |


## KubernetesServiceSpec



KubernetesServiceSpec defines the desired state of the Kubernetes service resource.



_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations that should be appended to the service.<br />By default, no annotations are appended. |  |  |
| `type` _[ServiceType](#servicetype)_ | Type determines how the Service is exposed. Defaults to LoadBalancer.<br />Valid options are ClusterIP and LoadBalancer.<br />"LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it).<br />"ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP. | LoadBalancer | Enum: [LoadBalancer ClusterIP] <br /> |


## LogComponent

_Underlying type:_ _string_

LogComponent defines a component that supports a configured logging level.
This type is not implemented until https://github.com/envoyproxy/gateway/issues/280
is fixed.

_Validation:_
- Enum: [system upstream http connection admin client filter main router runtime]

_Appears in:_
- [ProxyLogging](#proxylogging)

| Field | Description |
| --- | --- |
| `system` | LogComponentSystem defines the "system"-wide logging component. When specified,<br />all other logging components are ignored.<br /> |
| `upstream` | LogComponentUpstream defines defines the "upstream" logging component.<br /> |
| `http` | LogComponentHTTP defines defines the "http" logging component.<br /> |
| `connection` | LogComponentConnection defines defines the "connection" logging component.<br /> |
| `admin` | LogComponentAdmin defines defines the "admin" logging component.<br /> |
| `client` | LogComponentClient defines defines the "client" logging component.<br /> |
| `filter` | LogComponentFilter defines defines the "filter" logging component.<br /> |
| `main` | LogComponentMain defines defines the "main" logging component.<br /> |
| `router` | LogComponentRouter defines defines the "router" logging component.<br /> |
| `runtime` | LogComponentRuntime defines defines the "runtime" logging component.<br /> |


## LogLevel

_Underlying type:_ _string_

LogLevel defines a log level for system logs. This type is not implemented until
https://github.com/envoyproxy/gateway/issues/280 is fixed.

_Validation:_
- Enum: [debug info error]

_Appears in:_
- [ProxyLogging](#proxylogging)

| Field | Description |
| --- | --- |
| `debug` | LogLevelDebug defines the "debug" logging level.<br /> |
| `info` | LogLevelInfo defines the "Info" logging level.<br /> |
| `error` | LogLevelError defines the "Error" logging level.<br /> |




## ProviderType

_Underlying type:_ _string_

ProviderType defines the types of providers supported by Envoy Gateway.

_Validation:_
- Enum: [Kubernetes]

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)
- [EnvoyProxyProvider](#envoyproxyprovider)

| Field | Description |
| --- | --- |
| `Kubernetes` | ProviderTypeKubernetes defines the "Kubernetes" provider.<br /> |
| `File` | ProviderTypeFile defines the "File" provider. This type is not implemented<br />until https://github.com/envoyproxy/gateway/issues/1001 is fixed.<br /> |


## ProxyLogging



ProxyLogging defines logging parameters for managed proxies. This type is not
implemented until https://github.com/envoyproxy/gateway/issues/280 is fixed.



_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `level` _object (keys:[LogComponent](#logcomponent), values:[LogLevel](#loglevel))_ | Level is a map of logging level per component, where the component is the key<br />and the log level is the value. If unspecified, defaults to "System: Info". | \{ system:info \} |  |


## RateLimit



RateLimit defines the configuration associated with the Rate Limit Service
used for Global Rate Limiting.



_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `backend` _[RateLimitDatabaseBackend](#ratelimitdatabasebackend)_ | Backend holds the configuration associated with the<br />database backend used by the rate limit service to store<br />state associated with global ratelimiting. |  |  |


## RateLimitDatabaseBackend



RateLimitDatabaseBackend defines the configuration associated with
the database backend used by the rate limit service.



_Appears in:_
- [RateLimit](#ratelimit)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[RateLimitDatabaseBackendType](#ratelimitdatabasebackendtype)_ | Type is the type of database backend to use. Supported types are:<br />	* Redis: Connects to a Redis database. |  | Enum: [Redis] <br /> |
| `redis` _[RateLimitRedisSettings](#ratelimitredissettings)_ | Redis defines the settings needed to connect to a Redis database. |  |  |


## RateLimitDatabaseBackendType

_Underlying type:_ _string_

RateLimitDatabaseBackendType specifies the types of database backend
to be used by the rate limit service.

_Validation:_
- Enum: [Redis]

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Description |
| --- | --- |
| `Redis` | RedisBackendType uses a redis database for the rate limit service.<br /> |


## RateLimitRedisSettings



RateLimitRedisSettings defines the configuration for connecting to
a Redis database.



_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | URL of the Redis Database. |  |  |


## ServiceType

_Underlying type:_ _string_

ServiceType string describes ingress methods for a service

_Validation:_
- Enum: [LoadBalancer ClusterIP]

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)

| Field | Description |
| --- | --- |
| `LoadBalancer` | ServiceTypeLoadBalancer means a service will be exposed via an<br />external load balancer (if the cloud provider supports it).<br /> |
| `ClusterIP` | ServiceTypeClusterIP means a service will only be accessible inside the<br />cluster, via the cluster IP.<br /> |


## XDSTranslatorHook

_Underlying type:_ _string_

XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support
for the xds-translator

_Validation:_
- Enum: [VirtualHost Route HTTPListener Translation]

_Appears in:_
- [XDSTranslatorHooks](#xdstranslatorhooks)

| Field | Description |
| --- | --- |
| `VirtualHost` |  |
| `Route` |  |
| `HTTPListener` |  |
| `Translation` |  |


## XDSTranslatorHooks



XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.



_Appears in:_
- [ExtensionHooks](#extensionhooks)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `pre` _[XDSTranslatorHook](#xdstranslatorhook) array_ |  |  | Enum: [VirtualHost Route HTTPListener Translation] <br /> |
| `post` _[XDSTranslatorHook](#xdstranslatorhook) array_ |  |  | Enum: [VirtualHost Route HTTPListener Translation] <br /> |


