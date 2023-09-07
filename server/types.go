package server

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type ResourceListWrapper struct {
	List runtime.Object
}

func (rw *ResourceListWrapper) Update(newList runtime.Object) {
	rw.List = newList
}

// ResourceInterface provides generic operations for Kubernetes resources
type ResourceInterface interface {
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type ClusterRoleBindingAdapter struct {
	client *kubernetes.Clientset
}

func (c *ClusterRoleBindingAdapter) List(opts metav1.ListOptions) (runtime.Object, error) {
	return c.client.RbacV1().ClusterRoleBindings().List(context.Background(), opts)
}

func (c *ClusterRoleBindingAdapter) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.RbacV1().ClusterRoleBindings().Watch(context.Background(), opts)
}

// RoleBindingAdapter provides operations for RoleBinding resources
type RoleBindingAdapter struct {
	client *kubernetes.Clientset
}

func (r *RoleBindingAdapter) List(opts metav1.ListOptions) (runtime.Object, error) {
	return r.client.RbacV1().RoleBindings(metav1.NamespaceAll).List(context.Background(), opts)
}

func (r *RoleBindingAdapter) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return r.client.RbacV1().RoleBindings(metav1.NamespaceAll).Watch(context.Background(), opts)
}

// RoleAdapter provides operations for Roles resources
type RoleAdapter struct {
	client *kubernetes.Clientset
}

func (r *RoleAdapter) List(opts metav1.ListOptions) (runtime.Object, error) {
	return r.client.RbacV1().Roles(metav1.NamespaceAll).List(context.Background(), opts)
}

func (r *RoleAdapter) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return r.client.RbacV1().Roles(metav1.NamespaceAll).Watch(context.Background(), opts)
}

// ClusterRoleAdapter provides operations for ClusterRoles resources
type ClusterRoleAdapter struct {
	client *kubernetes.Clientset
}

func (c *ClusterRoleAdapter) List(opts metav1.ListOptions) (runtime.Object, error) {
	return c.client.RbacV1().ClusterRoles().List(context.Background(), opts)
}

func (c *ClusterRoleAdapter) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.RbacV1().ClusterRoles().Watch(context.Background(), opts)
}
