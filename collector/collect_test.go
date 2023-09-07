package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1r "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHasPermission(t *testing.T) {
	rules := []v1r.PolicyRule{
		{
			Verbs: []string{"get", "list"},
		},
	}
	assert.True(t, hasPermission(rules))

	rules = []v1r.PolicyRule{
		{
			Verbs: []string{"delete"},
		},
	}
	assert.False(t, hasPermission(rules))
}

func TestGetRoles(t *testing.T) {
	roleList := v1r.RoleList{
		Items: []v1r.Role{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testRole",
				},
				Rules: []v1r.PolicyRule{
					{
						Verbs: []string{"get", "list"},
					},
				},
			},
		},
	}

	clusterRoleList := v1r.ClusterRoleList{
		Items: []v1r.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testClusterRole",
				},
				Rules: []v1r.PolicyRule{
					{
						Verbs: []string{"delete"},
					},
				},
			},
		},
	}

	rolesWithPerm, clusterRolesWithPerm := GetRoles(roleList, clusterRoleList)
	assert.True(t, rolesWithPerm["testRole"])
	assert.False(t, clusterRolesWithPerm["testClusterRole"])
}

func TestCollectOutput(t *testing.T) {
	out := make(chan RBACCollect, 1000)
	out <- RBACCollect{subject: "test", namespace: "namespace1"}
	out <- RBACCollect{subject: "test", namespace: "namespace2"}
	close(out)

	perms := CollectOutput(out)
	assert.True(t, perms["test"]["namespace1"])
	assert.True(t, perms["test"]["namespace2"])
}

func TestCollect(t *testing.T) {
	tests := []struct {
		name                string
		roles               map[string]bool
		clusterRoles        map[string]bool
		roleBindings        v1r.RoleBindingList
		clusterRoleBindings v1r.ClusterRoleBindingList
		expectedOutput      map[string]map[string]bool
	}{
		{
			name: "basic test",
			roles: map[string]bool{
				"testRole": true,
			},
			clusterRoles: map[string]bool{
				"testClusterRole": true,
			},
			roleBindings: v1r.RoleBindingList{
				Items: []v1r.RoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testRoleBinding",
							Namespace: "testNamespace",
						},
						RoleRef: v1r.RoleRef{
							Name: "testRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "User",
								Name: "testUser",
							},
						},
					},
				},
			},
			clusterRoleBindings: v1r.ClusterRoleBindingList{
				Items: []v1r.ClusterRoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testClusterRoleBinding",
						},
						RoleRef: v1r.RoleRef{
							Name: "testClusterRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "User",
								Name: "testClusterUser",
							},
						},
					},
				},
			},
			expectedOutput: map[string]map[string]bool{
				"testUser": {
					"testNamespace": true,
				},
				"testClusterUser": {
					"#cluster-wide": true,
				},
			},
		},
		{
			name: "role without permission",
			roles: map[string]bool{
				"noPermissionRole": false,
			},
			clusterRoles: map[string]bool{},
			roleBindings: v1r.RoleBindingList{
				Items: []v1r.RoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testRoleBinding",
							Namespace: "testNamespace",
						},
						RoleRef: v1r.RoleRef{
							Name: "noPermissionRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "User",
								Name: "testUser",
							},
						},
					},
				},
			},
			clusterRoleBindings: v1r.ClusterRoleBindingList{},
			expectedOutput:      map[string]map[string]bool{},
		},
		{
			name: "system user in roleBinding",
			roles: map[string]bool{
				"testRole": true,
			},
			clusterRoles: map[string]bool{},
			roleBindings: v1r.RoleBindingList{
				Items: []v1r.RoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testRoleBinding",
							Namespace: "testNamespace",
						},
						RoleRef: v1r.RoleRef{
							Name: "testRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "User",
								Name: "system:testUser",
							},
						},
					},
				},
			},
			clusterRoleBindings: v1r.ClusterRoleBindingList{},
			expectedOutput:      map[string]map[string]bool{},
		},
		{
			name:  "service account in clusterRoleBinding",
			roles: map[string]bool{},
			clusterRoles: map[string]bool{
				"testClusterRole": true,
			},
			roleBindings: v1r.RoleBindingList{},
			clusterRoleBindings: v1r.ClusterRoleBindingList{
				Items: []v1r.ClusterRoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testClusterRoleBinding",
						},
						RoleRef: v1r.RoleRef{
							Name: "testClusterRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "ServiceAccount",
								Name: "testSA",
							},
						},
					},
				},
			},
			expectedOutput: map[string]map[string]bool{},
		},
		{
			name:  "cluster-wide permission",
			roles: map[string]bool{},
			clusterRoles: map[string]bool{
				"testClusterRole": true,
			},
			roleBindings: v1r.RoleBindingList{},
			clusterRoleBindings: v1r.ClusterRoleBindingList{
				Items: []v1r.ClusterRoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testClusterRoleBinding",
						},
						RoleRef: v1r.RoleRef{
							Name: "testClusterRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "User",
								Name: "testUser",
							},
						},
					},
				},
			},
			expectedOutput: map[string]map[string]bool{
				"testUser": {
					"#cluster-wide": true,
				},
			},
		},
		{
			name: "Multiple users in a single roleBinding",
			roles: map[string]bool{
				"multiUserRole": true,
			},
			clusterRoles: map[string]bool{},
			roleBindings: v1r.RoleBindingList{
				Items: []v1r.RoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "multiUserRoleBinding",
							Namespace: "multiUserNamespace",
						},
						RoleRef: v1r.RoleRef{
							Name: "multiUserRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "User",
								Name: "userA",
							},
							{
								Kind: "User",
								Name: "userB",
							},
						},
					},
				},
			},
			clusterRoleBindings: v1r.ClusterRoleBindingList{},
			expectedOutput: map[string]map[string]bool{
				"userA": {
					"multiUserNamespace": true,
				},
				"userB": {
					"multiUserNamespace": true,
				},
			},
		},
		{
			name: "RoleBinding with ServiceAccount and non-System User",
			roles: map[string]bool{
				"mixedRole": true,
			},
			clusterRoles: map[string]bool{},
			roleBindings: v1r.RoleBindingList{
				Items: []v1r.RoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mixedRoleBinding",
							Namespace: "mixedNamespace",
						},
						RoleRef: v1r.RoleRef{
							Name: "mixedRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "ServiceAccount",
								Name: "serviceAccountA",
							},
							{
								Kind: "User",
								Name: "userC",
							},
						},
					},
				},
			},
			clusterRoleBindings: v1r.ClusterRoleBindingList{},
			expectedOutput: map[string]map[string]bool{
				"userC": {
					"mixedNamespace": true,
				},
			},
		},
		{
			name: "RoleBinding with non-system service account",
			roles: map[string]bool{
				"saRole": true,
			},
			clusterRoles: map[string]bool{},
			roleBindings: v1r.RoleBindingList{
				Items: []v1r.RoleBinding{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "saRoleBinding",
							Namespace: "saNamespace",
						},
						RoleRef: v1r.RoleRef{
							Name: "saRole",
						},
						Subjects: []v1r.Subject{
							{
								Kind: "ServiceAccount",
								Name: "nonSystemSA",
							},
						},
					},
				},
			},
			clusterRoleBindings: v1r.ClusterRoleBindingList{},
			expectedOutput:      map[string]map[string]bool{},
		},
		{
			name:  "clusterRole with no associated bindings",
			roles: map[string]bool{},
			clusterRoles: map[string]bool{
				"orphanClusterRole": true,
			},
			roleBindings:        v1r.RoleBindingList{},
			clusterRoleBindings: v1r.ClusterRoleBindingList{},
			expectedOutput:      map[string]map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Collect(tt.roles, tt.clusterRoles, &tt.roleBindings, &tt.clusterRoleBindings)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}
