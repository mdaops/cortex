package main

import (
	"context"

	"github.com/mdaops/cortex/configurations/pkg/resources"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

// Function implements the Crossplane composition function for Tenant resources.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// tenantComposer holds state for composing tenant resources.
type tenantComposer struct {
	oxr     *resource.Composite
	name    string
	desired map[resource.Name]any
}

func newTenantComposer(oxr *resource.Composite) (*tenantComposer, error) {
	name, err := oxr.Resource.GetString("spec.name")
	if err != nil {
		return nil, errors.Wrap(err, "spec.name is required")
	}
	return &tenantComposer{
		oxr:     oxr,
		name:    name,
		desired: make(map[resource.Name]any),
	}, nil
}

func (tc *tenantComposer) getString(path string) string {
	val, _ := tc.oxr.Resource.GetString(path)
	return val
}

func (tc *tenantComposer) getStringArray(path string) []string {
	val, _ := tc.oxr.Resource.GetStringArray(path)
	return val
}

func (tc *tenantComposer) featureEnabled(feature string) bool {
	enabled, err := tc.oxr.Resource.GetBool("spec.features." + feature + ".enabled")
	return err == nil && enabled
}

func (tc *tenantComposer) add(name resource.Name, obj any) {
	tc.desired[name] = obj
}

func (tc *tenantComposer) composeBase() error {
	tc.add("namespace", resources.NewTenantNamespace(tc.name))

	project, err := resources.NewTenantProject(tc.name, tc.getString("spec.description"), tc.getStringArray("spec.sourceRepos"))
	if err != nil {
		return errors.Wrap(err, "spec.sourceRepos is required")
	}
	tc.add("project", project)
	return nil
}

func (tc *tenantComposer) composeArgoWorkflows() {
	if !tc.featureEnabled("argoWorkflows") {
		return
	}
	tc.add("argo-workflow-sa", resources.NewServiceAccount(resources.ServiceAccountConfig{
		Name:      "argo-workflow",
		Namespace: tc.name,
	}))
	tc.add("argo-workflow-role", resources.NewRole(resources.RoleConfig{
		Name:      "argo-workflow",
		Namespace: tc.name,
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"workflowtaskresults"},
			Verbs:     []string{"create", "patch"},
		}},
	}))
	tc.add("argo-workflow-rolebinding", resources.NewRoleBinding(resources.RoleBindingConfig{
		Name:               "argo-workflow",
		Namespace:          tc.name,
		RoleName:           "argo-workflow",
		ServiceAccountName: "argo-workflow",
	}))
}

// RunFunction composes namespace and ArgoCD project resources for a Tenant.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	rsp := response.To(req, response.DefaultTTL)

	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed composite resource"))
		return rsp, nil
	}

	log := f.log.WithValues(
		"xr-name", oxr.Resource.GetName(),
		"xr-kind", oxr.Resource.GetKind(),
	)

	tc, err := newTenantComposer(oxr)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	if err := tc.composeBase(); err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	tc.composeArgoWorkflows()

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
		return rsp, nil
	}

	for name, obj := range tc.desired {
		c := composed.New()
		if err := resources.ConvertViaJSON(c, obj); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot convert %s to unstructured", name))
			return rsp, nil
		}
		desired[name] = &resource.DesiredComposed{Resource: c}
	}

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	log.Info("Composed tenant resources", "tenant", tc.name)
	response.ConditionTrue(rsp, "FunctionSuccess", "Success").TargetCompositeAndClaim()

	return rsp, nil
}
