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

// composer holds state for composing resources from an XR.
type composer struct {
	oxr     *resource.Composite
	desired map[resource.Name]any
}

func newComposer(oxr *resource.Composite) *composer {
	return &composer{
		oxr:     oxr,
		desired: make(map[resource.Name]any),
	}
}

func (c *composer) getString(path string) string {
	val, _ := c.oxr.Resource.GetString(path)
	return val
}

func (c *composer) getStringArray(path string) []string {
	val, _ := c.oxr.Resource.GetStringArray(path)
	return val
}

func (c *composer) featureEnabled(feature string) bool {
	enabled, err := c.oxr.Resource.GetBool("spec.features." + feature + ".enabled")
	return err == nil && enabled
}

func (c *composer) add(name resource.Name, obj any) {
	c.desired[name] = obj
}

func (c *composer) name() string {
	return c.getString("spec.name")
}

func (c *composer) composeBase() error {
	name := c.name()
	if name == "" {
		return errors.New("spec.name is required")
	}

	c.add("namespace", resources.NewTenantNamespace(name))

	project, err := resources.NewTenantProject(name, c.getString("spec.description"), c.getStringArray("spec.sourceRepos"))
	if err != nil {
		return errors.Wrap(err, "spec.sourceRepos is required")
	}
	c.add("project", project)
	return nil
}

func (c *composer) composeArgoWorkflows() {
	if !c.featureEnabled("argoWorkflows") {
		return
	}
	name := c.name()
	c.add("argo-workflow-sa", resources.NewServiceAccount(resources.ServiceAccountConfig{
		Name:      "argo-workflow",
		Namespace: name,
	}))
	c.add("argo-workflow-role", resources.NewRole(resources.RoleConfig{
		Name:      "argo-workflow",
		Namespace: name,
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"workflowtaskresults"},
			Verbs:     []string{"create", "patch"},
		}},
	}))
	c.add("argo-workflow-rolebinding", resources.NewRoleBinding(resources.RoleBindingConfig{
		Name:               "argo-workflow",
		Namespace:          name,
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

	c := newComposer(oxr)

	if err := c.composeBase(); err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	c.composeArgoWorkflows()

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
		return rsp, nil
	}

	for name, obj := range c.desired {
		res := composed.New()
		if err := resources.ConvertViaJSON(res, obj); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot convert %s to unstructured", name))
			return rsp, nil
		}
		desired[name] = &resource.DesiredComposed{Resource: res}
	}

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	log.Info("Composed tenant resources", "tenant", c.name())
	response.ConditionTrue(rsp, "FunctionSuccess", "Success").TargetCompositeAndClaim()

	return rsp, nil
}
