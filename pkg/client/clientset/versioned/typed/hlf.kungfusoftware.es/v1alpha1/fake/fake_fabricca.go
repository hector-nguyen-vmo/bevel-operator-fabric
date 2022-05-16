/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/hyperledger-labs/hlf-operator/api/hlf.kungfusoftware.es/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFabricCAs implements FabricCAInterface
type FakeFabricCAs struct {
	Fake *FakeHlfV1alpha1
	ns   string
}

var fabriccasResource = schema.GroupVersionResource{Group: "hlf.kungfusoftware.es", Version: "v1alpha1", Resource: "fabriccas"}

var fabriccasKind = schema.GroupVersionKind{Group: "hlf.kungfusoftware.es", Version: "v1alpha1", Kind: "FabricCA"}

// Get takes name of the fabricCA, and returns the corresponding fabricCA object, and an error if there is any.
func (c *FakeFabricCAs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.FabricCA, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(fabriccasResource, c.ns, name), &v1alpha1.FabricCA{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FabricCA), err
}

// List takes label and field selectors, and returns the list of FabricCAs that match those selectors.
func (c *FakeFabricCAs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FabricCAList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(fabriccasResource, fabriccasKind, c.ns, opts), &v1alpha1.FabricCAList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.FabricCAList{ListMeta: obj.(*v1alpha1.FabricCAList).ListMeta}
	for _, item := range obj.(*v1alpha1.FabricCAList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested fabricCAs.
func (c *FakeFabricCAs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(fabriccasResource, c.ns, opts))

}

// Create takes the representation of a fabricCA and creates it.  Returns the server's representation of the fabricCA, and an error, if there is any.
func (c *FakeFabricCAs) Create(ctx context.Context, fabricCA *v1alpha1.FabricCA, opts v1.CreateOptions) (result *v1alpha1.FabricCA, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(fabriccasResource, c.ns, fabricCA), &v1alpha1.FabricCA{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FabricCA), err
}

// Update takes the representation of a fabricCA and updates it. Returns the server's representation of the fabricCA, and an error, if there is any.
func (c *FakeFabricCAs) Update(ctx context.Context, fabricCA *v1alpha1.FabricCA, opts v1.UpdateOptions) (result *v1alpha1.FabricCA, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(fabriccasResource, c.ns, fabricCA), &v1alpha1.FabricCA{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FabricCA), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFabricCAs) UpdateStatus(ctx context.Context, fabricCA *v1alpha1.FabricCA, opts v1.UpdateOptions) (*v1alpha1.FabricCA, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(fabriccasResource, "status", c.ns, fabricCA), &v1alpha1.FabricCA{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FabricCA), err
}

// Delete takes name of the fabricCA and deletes it. Returns an error if one occurs.
func (c *FakeFabricCAs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(fabriccasResource, c.ns, name), &v1alpha1.FabricCA{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFabricCAs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(fabriccasResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.FabricCAList{})
	return err
}

// Patch applies the patch and returns the patched fabricCA.
func (c *FakeFabricCAs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.FabricCA, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(fabriccasResource, c.ns, name, pt, data, subresources...), &v1alpha1.FabricCA{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FabricCA), err
}
