package main

import (
	"context"
	"fmt"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

// Function composes Tenant resources.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	rsp := response.To(req, response.DefaultTTL)

	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	log := f.log.WithValues(
		"xr-name", oxr.Resource.GetName(),
		"xr-kind", oxr.Resource.GetKind(),
	)

	// Get tenant name from spec
	tenantName, err := oxr.Resource.GetString("spec.name")
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get spec.name"))
		return rsp, nil
	}

	// Get optional fields with defaults
	description, _ := oxr.Resource.GetString("spec.description")
	if description == "" {
		description = fmt.Sprintf("%s tenant workloads", tenantName)
	}

	sourceRepos, _ := oxr.Resource.GetStringArray("spec.sourceRepos")
	if len(sourceRepos) == 0 {
		sourceRepos = []string{"*"}
	}

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	}

	// Create Namespace via provider-kubernetes Object
	namespace := composed.New()
	namespace.SetAPIVersion("kubernetes.crossplane.io/v1alpha2")
	namespace.SetKind("Object")
	namespace.SetName(fmt.Sprintf("%s-namespace", tenantName))
	_ = namespace.SetValue("spec.forProvider.manifest", map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": tenantName,
			"labels": map[string]interface{}{
				"platform.synapse.io/tenant": tenantName,
			},
		},
	})
	_ = namespace.SetValue("spec.providerConfigRef.name", "default")
	desired[resource.Name("namespace")] = &resource.DesiredComposed{Resource: namespace}

	// Create ArgoCD AppProject via provider-argocd
	project := composed.New()
	project.SetAPIVersion("projects.argocd.crossplane.io/v1alpha1")
	project.SetKind("Project")
	project.SetName(fmt.Sprintf("%s-project", tenantName))
	_ = project.SetValue("spec.forProvider.metadata.name", tenantName)
	_ = project.SetValue("spec.forProvider.metadata.namespace", "argo-system")
	_ = project.SetValue("spec.forProvider.description", description)
	_ = project.SetValue("spec.forProvider.sourceRepos", sourceRepos)
	_ = project.SetValue("spec.forProvider.destinations", []interface{}{
		map[string]interface{}{
			"namespace": tenantName,
			"server":    "https://kubernetes.default.svc",
		},
		map[string]interface{}{
			"namespace": fmt.Sprintf("%s-*", tenantName),
			"server":    "https://kubernetes.default.svc",
		},
	})
	_ = project.SetValue("spec.providerConfigRef.name", "default")
	desired[resource.Name("project")] = &resource.DesiredComposed{Resource: project}

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources in %T", rsp))
		return rsp, nil
	}

	log.Info("Composed tenant resources", "tenant", tenantName)
	response.ConditionTrue(rsp, "FunctionSuccess", "Success").
		TargetCompositeAndClaim()

	return rsp, nil
}
