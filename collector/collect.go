package collector

import (
	"strings"

	v1r "k8s.io/api/rbac/v1"
)

type RBACCollect struct {
	subject   string
	namespace string
}

func hasPermission(rules []v1r.PolicyRule) bool {
	for _, rule := range rules {
		for _, verb := range rule.Verbs {
			if verb == "*" || verb == "approve" || verb == "create" || verb == "edit" || verb == "escalate" || verb == "get" || verb == "impersonate" || verb == "list" || verb == "patch" || verb == "update" || verb == "use" || verb == "view" || verb == "watch" {
				return true
			}
		}
	}
	return false
}

func GetRoles(roles v1r.RoleList, clusterRoles v1r.ClusterRoleList) (map[string]bool, map[string]bool) {
	clusterRolesWithPerm := make(map[string]bool)
	rolesWithPerm := make(map[string]bool)

	for _, role := range roles.Items {
		rolesWithPerm[role.Name] = hasPermission(role.Rules)
	}

	for _, clusterRole := range clusterRoles.Items {
		perm := hasPermission(clusterRole.Rules)
		clusterRolesWithPerm[clusterRole.Name] = perm
		rolesWithPerm[clusterRole.Name] = perm
	}

	return rolesWithPerm, clusterRolesWithPerm
}

func CollectOutput(out chan RBACCollect) map[string]map[string]bool {
	permissions := make(map[string]map[string]bool, 1000)
	for o := range out {
		if _, ok := permissions[o.subject]; !ok {
			permissions[o.subject] = make(map[string]bool, 1000)
		}
		permissions[o.subject][o.namespace] = true
	}
	return permissions
}

func Collect(roles, clusterRoles map[string]bool, roleBindings *v1r.RoleBindingList, clusterRoleBindings *v1r.ClusterRoleBindingList) map[string]map[string]bool {
	out := make(chan RBACCollect, 1000)
	for _, rb := range roleBindings.Items {
		if roles[rb.RoleRef.Name] {
			for _, subject := range rb.Subjects {
				if subject.Kind == "ServiceAccount" || rb.Namespace == "openshift-console-user-settings" || strings.Contains(subject.Name, "system") {
					continue
				}
				out <- RBACCollect{subject: subject.Name, namespace: rb.Namespace}
			}
		}
	}

	for _, crb := range clusterRoleBindings.Items {
		if clusterRoles[crb.RoleRef.Name] {
			for _, subject := range crb.Subjects {
				if subject.Kind == "ServiceAccount" || strings.Contains(subject.Name, "system") {
					continue
				}
				out <- RBACCollect{subject: subject.Name, namespace: "#cluster-wide"}
			}
		}
	}
	close(out)
	return CollectOutput(out)
}
