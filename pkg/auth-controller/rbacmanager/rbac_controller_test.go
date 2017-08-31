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

package rbacmanager

import (
	"fmt"
	"reflect"
	"testing"

	crv1 "git.openstack.org/openstack/stackube/pkg/apis/v1"
	"git.openstack.org/openstack/stackube/pkg/auth-controller/rbacmanager/rbac"
	crdClient "git.openstack.org/openstack/stackube/pkg/kubecrd"
	"git.openstack.org/openstack/stackube/pkg/util"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	userCIDR    = "10.244.0.0/16"
	userGateway = "10.244.0.1"
)

var systemTenant = &crv1.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name:      util.SystemTenant,
		Namespace: util.SystemTenant,
	},
	Spec: crv1.TenantSpec{
		UserName: util.SystemTenant,
		Password: util.SystemPassword,
	},
}

var systemNetwork = &crv1.Network{
	ObjectMeta: metav1.ObjectMeta{
		Name:      util.SystemNetwork,
		Namespace: util.SystemTenant,
	},
	Spec: crv1.NetworkSpec{
		CIDR:    userCIDR,
		Gateway: userGateway,
	},
}

func newNetworkForTenant(namespace string) *crv1.Network {
	return &crv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace,
			Namespace: namespace,
		},
		Spec: crv1.NetworkSpec{
			CIDR:    userCIDR,
			Gateway: userGateway,
		},
	}
}

func newNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newController() (*Controller, *crdClient.FakeCRDClient, *fake.Clientset, error) {
	client := fake.NewSimpleClientset()

	kubeCRDClient, err := crdClient.NewFake()
	if err != nil {
		return nil, nil, nil, err
	}

	controller, _ := NewRBACController(client, kubeCRDClient, userCIDR, userGateway)

	return controller, kubeCRDClient, client, nil
}

func TestCreateNetworkForTenant(t *testing.T) {
	testNamespace := "test"
	var controller *Controller
	var kubeCRDClient *crdClient.FakeCRDClient
	var err error

	testCases := []struct {
		testName  string
		updateFn  func() error
		expectErr bool
	}{
		{
			testName: "Failed add network",
			updateFn: func() error {

				// Creates a new fake controller
				controller, kubeCRDClient, _, err = newController()
				if err != nil {
					t.Fatalf("Failed start a new fake controller: %v", err)
				}
				// Injects AddNetwork error
				kubeCRDClient.InjectError("AddNetwork", fmt.Errorf("failed to create Network"))
				return controller.createNetworkForTenant(testNamespace)
			},
			expectErr: false,
		},
		{
			testName: "Success",
			updateFn: func() error {

				// Create a new fake controller.
				controller, _, _, err = newController()
				if err != nil {
					t.Fatalf("Failed start a new fake controller: %v", err)
				}
				return controller.createNetworkForTenant("test")
			},
			expectErr: true,
		},
	}

	for tci, tc := range testCases {
		err := tc.updateFn()
		if !tc.expectErr && err == nil {
			t.Errorf("Case[%d]: %s expected error, got nil", tci, tc.testName)
		} else if tc.expectErr && err != nil {
			t.Errorf("Case[%d]: %s expected success, got error %v", tci, tc.testName, err)

			if len(kubeCRDClient.Networks) != 1 {
				t.Errorf("Case[%d]: %s expected 1 networks to be created, got %v", tci, tc.testName, kubeCRDClient.Networks)
			}
			network, ok := kubeCRDClient.Networks[testNamespace]
			if !ok {
				t.Errorf("Case[%d]: %s expected %s network to be created, got none", tci, tc.testName, testNamespace)
			} else if reflect.DeepEqual(network, newNetworkForTenant(testNamespace)) {
				t.Errorf("Case[%d]: %s expected the created %s network has incorrect parameters: %v", tci, tc.testName, testNamespace, network)
			}
		}
	}
}

func TestInitSystemReservedTenantNetwork(t *testing.T) {
	var controller *Controller
	var kubeCRDClient *crdClient.FakeCRDClient
	var err error

	testCases := []struct {
		testName  string
		updateFn  func() error
		expectErr bool
	}{
		{
			testName: "Failed add tenant",
			updateFn: func() error {

				// Create a new fake controller.
				controller, kubeCRDClient, _, err = newController()
				if err != nil {
					t.Fatalf("Failed start a new fake controller: %v", err)
				}
				// Injects AddTenant error.
				kubeCRDClient.InjectError("AddTenant", fmt.Errorf("failed to create Tenant"))
				return controller.initSystemReservedTenantNetwork()

			},
			expectErr: false,
		},
		{
			testName: "Failed add network",
			updateFn: func() error {

				// Create a new fake controller.
				controller, kubeCRDClient, _, err = newController()
				if err != nil {
					t.Fatalf("Failed start a new fake controller: %v", err)
				}
				// Inject AddNetwork error
				kubeCRDClient.InjectError("AddNetwork", fmt.Errorf("failed to create Network"))
				return controller.initSystemReservedTenantNetwork()
			},
			expectErr: false,
		},
		{
			testName: "Success",
			updateFn: func() error {

				// Create a new fake controller.
				controller, _, _, err = newController()
				if err != nil {
					t.Fatalf("Failed start a new fake controller: %v", err)
				}
				return controller.initSystemReservedTenantNetwork()
			},
			expectErr: true,
		},
	}

	for tci, tc := range testCases {
		err := tc.updateFn()
		if !tc.expectErr && err == nil {
			t.Errorf("Case[%d]: %v expected error, got nil", tci, tc.testName)
		} else if tc.expectErr && err != nil {
			t.Errorf("Case[%d]: %v expected success, got error %v", tci, tc.testName, err)
			if len(kubeCRDClient.Tenants) != 1 {
				t.Errorf("Expected 1 tenants to be created, got %v", kubeCRDClient.Tenants)
			}
			if len(kubeCRDClient.Networks) != 1 {
				t.Errorf("Expected 1 networks to be created, got %v", kubeCRDClient.Networks)
			}

			tenant, ok := kubeCRDClient.Tenants["default"]
			if !ok {
				t.Errorf("Expected default tenant to be created, got none")
			} else if reflect.DeepEqual(tenant, systemTenant) {
				t.Errorf("The created default tenant has incorrect parameters: %v", tenant)
			}

			network, ok := kubeCRDClient.Networks["default"]
			if !ok {
				t.Errorf("Expected default network to be created, got none")
			} else if reflect.DeepEqual(network, systemNetwork) {
				t.Errorf("The created default network has incorrect parameters: %v", network)
			}
		}
	}
}

func testRBAC(t *testing.T, client *fake.Clientset, namespace string) {
	roleBinding, err := client.Rbac().RoleBindings(namespace).Get(namespace+"-rolebinding", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed get roleBindings: %v", err)
	}

	if !reflect.DeepEqual(roleBinding, rbac.GenerateRoleBinding(namespace, namespace)) {
		t.Errorf("Created rolebinding has incorrect parameters: %v", roleBinding)
	}

	saroleBinding, err := client.Rbac().RoleBindings(namespace).Get(namespace+"-rolebinding-sa", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed get ServiceAccount roleBindings: %v", err)
	}

	if !reflect.DeepEqual(saroleBinding, rbac.GenerateServiceAccountRoleBinding(namespace, namespace)) {
		t.Errorf("Created rolebinding has incorrect parameters: %v", saroleBinding)
	}

}

func TestSyncRBAC(t *testing.T) {
	testNamespace := "test"
	// Create a new fake controller.
	controller, _, client, err := newController()
	if err != nil {
		t.Fatalf("Failed start a new fake controller: %v", err)
	}
	ns := newNamespace(testNamespace)
	controller.syncRBAC(ns)
	testRBAC(t, client, testNamespace)
}

func TestOnAdd(t *testing.T) {
	var controller *Controller
	var kubeCRDClient *crdClient.FakeCRDClient
	var client *fake.Clientset
	var err error

	testCases := []struct {
		namespace  string
		updateFn   func(namespace string)
		expectedFn func(namespace string)
	}{
		{
			namespace: "default",
			updateFn: func(namespace string) {

				// Create a new fake controller.
				controller, kubeCRDClient, client, err = newController()
				if err != nil {
					t.Fatalf("Failed start a new fake controller: %v", err)
				}
				ns := newNamespace(namespace)
				controller.onAdd(ns)

			},
			expectedFn: func(namespace string) {
				if len(kubeCRDClient.Tenants) != 1 {
					t.Errorf("Expected 1 tenants to be created, got %v", kubeCRDClient.Tenants)
				}
				if len(kubeCRDClient.Networks) != 1 {
					t.Errorf("Expected 1 networks to be created, got %v", kubeCRDClient.Networks)
				}

				tenant, ok := kubeCRDClient.Tenants["default"]
				if !ok {
					t.Errorf("Expected default tenant to be created, got none")
				} else if !reflect.DeepEqual(tenant, systemTenant) {
					t.Errorf("The created default tenant has incorrect parameters: %v", tenant)
				}

				network, ok := kubeCRDClient.Networks["default"]
				if !ok {
					t.Errorf("Expected default network to be created, got none")
				} else if !reflect.DeepEqual(network, systemNetwork) {
					t.Errorf("The created default network has incorrect parameters: %v", network)
				}

				testRBAC(t, client, namespace)

			},
		},
		{
			namespace: "kube-system",
			updateFn: func(namespace string) {
				ns := newNamespace(namespace)
				controller.onAdd(ns)
			},
			expectedFn: func(namespace string) {
				if len(kubeCRDClient.Tenants) != 1 {
					t.Errorf("Expected 1 tenants to be created, got %v", kubeCRDClient.Tenants)
				}
				if len(kubeCRDClient.Networks) != 1 {
					t.Errorf("Expected 1 networks to be created, got %v", kubeCRDClient.Networks)
				}

				tenant, ok := kubeCRDClient.Tenants["default"]
				if !ok {
					t.Errorf("Expected default tenant to be created, got none")
				} else if !reflect.DeepEqual(tenant, systemTenant) {
					t.Errorf("The created default tenant has incorrect parameters: %v", tenant)
				}

				network, ok := kubeCRDClient.Networks["default"]
				if !ok {
					t.Errorf("Expected default network to be created, got none")
				} else if !reflect.DeepEqual(network, systemNetwork) {
					t.Errorf("The created default network has incorrect parameters: %v", network)
				}

				testRBAC(t, client, namespace)

			},
		},
		{
			namespace: "kube-public",
			updateFn: func(namespace string) {
				ns := newNamespace(namespace)
				controller.onAdd(ns)
			},
			expectedFn: func(namespace string) {
				if len(kubeCRDClient.Tenants) != 1 {
					t.Errorf("Expected 1 tenants to be created, got %v", kubeCRDClient.Tenants)
				}
				if len(kubeCRDClient.Networks) != 1 {
					t.Errorf("Expected 1 networks to be created, got %v", kubeCRDClient.Networks)
				}

				tenant, ok := kubeCRDClient.Tenants["default"]
				if !ok {
					t.Errorf("Expected default tenant to be created, got none")
				} else if !reflect.DeepEqual(tenant, systemTenant) {
					t.Errorf("The created default tenant has incorrect parameters: %v", tenant)
				}

				network, ok := kubeCRDClient.Networks["default"]
				if !ok {
					t.Errorf("Expected default network to be created, got none")
				} else if !reflect.DeepEqual(network, systemNetwork) {
					t.Errorf("The created default network has incorrect parameters: %v", network)
				}

				testRBAC(t, client, namespace)

			},
		},
		{
			namespace: "test",
			updateFn: func(namespace string) {
				ns := newNamespace(namespace)
				controller.onAdd(ns)

			},
			expectedFn: func(namespace string) {
				if len(kubeCRDClient.Tenants) != 1 {
					t.Errorf("Expected 1 tenants to be created, got %v", kubeCRDClient.Tenants)
				}
				if len(kubeCRDClient.Networks) != 2 {
					t.Errorf("Expected 2 networks to be created, got %v", kubeCRDClient.Networks)
				}

				tenant, ok := kubeCRDClient.Tenants["default"]
				if !ok {
					t.Errorf("Expected default tenant to be created, got none")
				} else if !reflect.DeepEqual(tenant, systemTenant) {
					t.Errorf("The created default tenant has incorrect parameters: %v", tenant)
				}

				network, ok := kubeCRDClient.Networks["default"]
				if !ok {
					t.Errorf("Expected default network to be created, got none")
				} else if !reflect.DeepEqual(network, systemNetwork) {
					t.Errorf("The created default network has incorrect parameters: %v", network)
				}

				network, ok = kubeCRDClient.Networks[namespace]
				if !ok {
					t.Errorf("Expected %s network to be created, got none", namespace)
				} else if !reflect.DeepEqual(network, newNetworkForTenant(namespace)) {
					t.Errorf("The created %s network has incorrect parameters: %v", namespace, network)
				}

				testRBAC(t, client, namespace)

			},
		},
	}

	for _, tc := range testCases {
		tc.updateFn(tc.namespace)
		tc.expectedFn(tc.namespace)
	}
}
