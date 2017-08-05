/*
Copyright (c) 2017 OpenStack Foundation.

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

package kubecrd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateNetworkCRD(t *testing.T) {
	clientset := apiextensionsclientfake.NewSimpleClientset()

	_, err := CreateNetworkCRD(clientset)
	assert.NoError(t, err)

	networkCRD, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(networkCRDName, metav1.GetOptions{})
	if err != nil {
		panic(fmt.Errorf("CustomResourceDefinitions.Create: %+v", err))
	}

	assert.Equal(t, networkCRDName, networkCRD.ObjectMeta.Name)
	assert.Equal(t, "networks", networkCRD.Spec.Names.Plural)
	assert.Equal(t, "network", networkCRD.Spec.Names.Singular)
	assert.Equal(t, "stackube.kubernetes.io", networkCRD.Spec.Group)
	assert.Equal(t, "v1", networkCRD.Spec.Version)
	assert.Equal(t, apiextensionsv1beta1.NamespaceScoped, networkCRD.Spec.Scope)
}
