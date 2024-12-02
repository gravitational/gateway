// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	gw := DefaultGateway()
	p := DefaultEnvoyGatewayProvider()
	return &EnvoyGateway{
		metav1.TypeMeta{
			Kind:       KindEnvoyGateway,
			APIVersion: GroupVersion.String(),
		},
		EnvoyGatewaySpec{
			Gateway:  gw,
			Provider: p,
		},
	}
}

// SetEnvoyGatewayDefaults sets default EnvoyGateway configuration parameters.
func (e *EnvoyGateway) SetEnvoyGatewayDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyGateway
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultEnvoyGatewayProvider()
	}
	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *Gateway {
	return &Gateway{
		ControllerName: GatewayControllerName,
	}
}

// DefaultEnvoyGatewayProvider returns a new EnvoyGatewayProvider with default configuration parameters.
func DefaultEnvoyGatewayProvider() *EnvoyGatewayProvider {
	return &EnvoyGatewayProvider{
		Type: ProviderTypeKubernetes,
	}
}

// GetEnvoyGatewayProvider returns the EnvoyGatewayProvider of EnvoyGateway or a default EnvoyGatewayProvider if unspecified.
func (e *EnvoyGateway) GetEnvoyGatewayProvider() *EnvoyGatewayProvider {
	if e.Provider != nil {
		return e.Provider
	}
	e.Provider = DefaultEnvoyGatewayProvider()

	return e.Provider
}

// DefaultEnvoyGatewayKubeProvider returns a new EnvoyGatewayKubernetesProvider with default settings.
func DefaultEnvoyGatewayKubeProvider() *EnvoyGatewayKubernetesProvider {
	return &EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: DefaultKubernetesDeployment(DefaultRateLimitImage),
	}
}

// DefaultEnvoyProxyProvider returns a new EnvoyProxyProvider with default settings.
func DefaultEnvoyProxyProvider() *EnvoyProxyProvider {
	return &EnvoyProxyProvider{
		Type: ProviderTypeKubernetes,
	}
}

// GetEnvoyProxyProvider returns the EnvoyProxyProvider of EnvoyProxy or a default EnvoyProxyProvider
// if unspecified.
func (e *EnvoyProxy) GetEnvoyProxyProvider() *EnvoyProxyProvider {
	if e.Spec.Provider != nil {
		return e.Spec.Provider
	}
	e.Spec.Provider = DefaultEnvoyProxyProvider()

	return e.Spec.Provider
}

// DefaultEnvoyProxyKubeProvider returns a new EnvoyProxyKubernetesProvider with default settings.
func DefaultEnvoyProxyKubeProvider() *EnvoyProxyKubernetesProvider {
	return &EnvoyProxyKubernetesProvider{
		EnvoyDeployment: DefaultKubernetesDeployment(DefaultEnvoyProxyImage),
		EnvoyService:    DefaultKubernetesService(),
	}
}

// DefaultKubernetesDeploymentReplicas returns the default replica settings.
func DefaultKubernetesDeploymentReplicas() *int32 {
	repl := int32(DefaultDeploymentReplicas)
	return &repl
}

// DefaultKubernetesContainerImage returns the default envoyproxy image.
func DefaultKubernetesContainerImage(image string) *string {
	return pointer.String(image)
}

// DefaultKubernetesDeployment returns a new KubernetesDeploymentSpec with default settings.
func DefaultKubernetesDeployment(image string) *KubernetesDeploymentSpec {
	return &KubernetesDeploymentSpec{
		Replicas:  DefaultKubernetesDeploymentReplicas(),
		Pod:       DefaultKubernetesPod(),
		Container: DefaultKubernetesContainer(image),
	}
}

// DefaultKubernetesPod returns a new KubernetesPodSpec with default settings.
func DefaultKubernetesPod() *KubernetesPodSpec {
	return &KubernetesPodSpec{}
}

// DefaultKubernetesContainer returns a new KubernetesContainerSpec with default settings.
func DefaultKubernetesContainer(image string) *KubernetesContainerSpec {
	return &KubernetesContainerSpec{
		Resources: DefaultResourceRequirements(),
		Image:     DefaultKubernetesContainerImage(image),
	}
}

// DefaultResourceRequirements returns a new ResourceRequirements with default settings.
func DefaultResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(DefaultDeploymentCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(DefaultDeploymentMemoryResourceRequests),
		},
	}
}

// DefaultKubernetesService returns a new KubernetesServiceSpec with default settings.
func DefaultKubernetesService() *KubernetesServiceSpec {
	return &KubernetesServiceSpec{
		Type: DefaultKubernetesServiceType(),
	}
}

// DefaultKubernetesServiceType returns a new KubernetesServiceType with default settings.
func DefaultKubernetesServiceType() *ServiceType {
	return GetKubernetesServiceType(ServiceTypeLoadBalancer)
}

// GetKubernetesServiceType returns the KubernetesServiceType pointer.
func GetKubernetesServiceType(serviceType ServiceType) *ServiceType {
	return &serviceType
}

// GetEnvoyProxyKubeProvider returns the EnvoyProxyKubernetesProvider of EnvoyProxyProvider or
// a default EnvoyProxyKubernetesProvider if unspecified. If EnvoyProxyProvider is not of
// type "Kubernetes", a nil EnvoyProxyKubernetesProvider is returned.
func (r *EnvoyProxyProvider) GetEnvoyProxyKubeProvider() *EnvoyProxyKubernetesProvider {
	if r.Type != ProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyProxyKubeProvider()
		return r.Kubernetes
	}

	if r.Kubernetes.EnvoyDeployment == nil {
		r.Kubernetes.EnvoyDeployment = DefaultKubernetesDeployment(DefaultEnvoyProxyImage)
	}

	if r.Kubernetes.EnvoyDeployment.Replicas == nil {
		r.Kubernetes.EnvoyDeployment.Replicas = DefaultKubernetesDeploymentReplicas()
	}

	if r.Kubernetes.EnvoyDeployment.Pod == nil {
		r.Kubernetes.EnvoyDeployment.Pod = DefaultKubernetesPod()
	}

	if r.Kubernetes.EnvoyDeployment.Container == nil {
		r.Kubernetes.EnvoyDeployment.Container = DefaultKubernetesContainer(DefaultEnvoyProxyImage)
	}

	if r.Kubernetes.EnvoyDeployment.Container.Resources == nil {
		r.Kubernetes.EnvoyDeployment.Container.Resources = DefaultResourceRequirements()
	}

	if r.Kubernetes.EnvoyDeployment.Container.Image == nil {
		r.Kubernetes.EnvoyDeployment.Container.Image = DefaultKubernetesContainerImage(DefaultEnvoyProxyImage)
	}

	if r.Kubernetes.EnvoyService == nil {
		r.Kubernetes.EnvoyService = DefaultKubernetesService()
	}

	if r.Kubernetes.EnvoyService.Type == nil {
		r.Kubernetes.EnvoyService.Type = GetKubernetesServiceType(ServiceTypeLoadBalancer)
	}

	return r.Kubernetes
}

// GetEnvoyGatewayKubeProvider returns the EnvoyGatewayKubernetesProvider of Provider or
// a default EnvoyGatewayKubernetesProvider if unspecified. If EnvoyGatewayProvider is not of
// type "Kubernetes", a nil EnvoyGatewayKubernetesProvider is returned.
func (r *EnvoyGatewayProvider) GetEnvoyGatewayKubeProvider() *EnvoyGatewayKubernetesProvider {
	if r.Type != ProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyGatewayKubeProvider()
		return r.Kubernetes
	}

	if r.Kubernetes.RateLimitDeployment == nil {
		r.Kubernetes.RateLimitDeployment = DefaultKubernetesDeployment(DefaultRateLimitImage)
	}

	if r.Kubernetes.RateLimitDeployment.Replicas == nil {
		r.Kubernetes.RateLimitDeployment.Replicas = DefaultKubernetesDeploymentReplicas()
	}

	if r.Kubernetes.RateLimitDeployment.Pod == nil {
		r.Kubernetes.RateLimitDeployment.Pod = DefaultKubernetesPod()
	}

	if r.Kubernetes.RateLimitDeployment.Container == nil {
		r.Kubernetes.RateLimitDeployment.Container = DefaultKubernetesContainer(DefaultRateLimitImage)
	}

	if r.Kubernetes.RateLimitDeployment.Container.Resources == nil {
		r.Kubernetes.RateLimitDeployment.Container.Resources = DefaultResourceRequirements()
	}

	if r.Kubernetes.RateLimitDeployment.Container.Image == nil {
		r.Kubernetes.RateLimitDeployment.Container.Image = DefaultKubernetesContainerImage(DefaultRateLimitImage)
	}

	return r.Kubernetes
}

// ApplyMergePatch applies a merge patch to a deployment based on the merge type
func (deployment *KubernetesDeploymentSpec) ApplyMergePatch(old *appv1.Deployment) (*appv1.Deployment, error) {
	if deployment.Patch == nil {
		return old, nil
	}
	var patchedJSON []byte
	var err error
	// Serialize the current deployment to JSON
	originalJSON, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("error marshaling original deployment: %w", err)
	}
	switch {
	case deployment.Patch.Type == nil || *deployment.Patch.Type == StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, deployment.Patch.Value.Raw, appv1.Deployment{})
	case *deployment.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, deployment.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *deployment.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}
	// Deserialize the patched JSON into a new deployment object
	var patchedDeployment appv1.Deployment
	if err := json.Unmarshal(patchedJSON, &patchedDeployment); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched deployment: %w", err)
	}
	return &patchedDeployment, nil
}