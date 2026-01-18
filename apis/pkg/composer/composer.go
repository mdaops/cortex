package composer

import (
	"errors"

	"github.com/crossplane/function-sdk-go/resource"
)

var ErrPathNotFound = errors.New("path not found")

// PathError wraps an error with the path that caused it.
type PathError struct {
	Path string
	Err  error
}

func (e *PathError) Error() string {
	return "path " + e.Path + ": " + e.Err.Error()
}

func (e *PathError) Unwrap() error {
	return e.Err
}

func (e *PathError) Is(target error) bool {
	return target == ErrPathNotFound
}

// Composer accumulates desired resources and errors from an XR.
type Composer struct {
	oxr     *resource.Composite
	desired map[resource.Name]any
	errs    []error
}

// New creates a Composer for the given observed composite resource.
func New(oxr *resource.Composite) *Composer {
	return &Composer{
		oxr:     oxr,
		desired: make(map[resource.Name]any),
	}
}

// GetString returns the string at path, recording an error if not found.
func (c *Composer) GetString(path string) string {
	val, err := c.oxr.Resource.GetString(path)
	if err != nil {
		c.errs = append(c.errs, &PathError{Path: path, Err: err})
		return ""
	}
	return val
}

// GetStringArray returns the string array at path, recording an error if not found.
func (c *Composer) GetStringArray(path string) []string {
	val, err := c.oxr.Resource.GetStringArray(path)
	if err != nil {
		c.errs = append(c.errs, &PathError{Path: path, Err: err})
		return nil
	}
	return val
}

// GetBool returns the bool at path, recording an error if not found.
func (c *Composer) GetBool(path string) bool {
	val, err := c.oxr.Resource.GetBool(path)
	if err != nil {
		c.errs = append(c.errs, &PathError{Path: path, Err: err})
		return false
	}
	return val
}

// Add adds a resource to the desired set.
func (c *Composer) Add(name resource.Name, obj any) {
	c.desired[name] = obj
}

// Desired returns the accumulated desired resources.
func (c *Composer) Desired() map[resource.Name]any {
	return c.desired
}

// Err returns accumulated errors, or nil if none.
func (c *Composer) Err() error {
	if len(c.errs) == 0 {
		return nil
	}
	return errors.Join(c.errs...)
}

// ClearErrs clears accumulated errors.
func (c *Composer) ClearErrs() {
	c.errs = nil
}
