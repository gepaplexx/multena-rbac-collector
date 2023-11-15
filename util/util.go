package util

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WriteConfigmap(clientset *kubernetes.Clientset, permission map[string]map[string]bool, c Config) error {
	permissions, err := yaml.Marshal(permission)
	if err != nil {
		return err
	}
	data := strings.Replace(string(permissions), "\n", "\\n", -1)
	patch := []byte(fmt.Sprintf(`{"data":{"labels.yaml": "%s"}}`, data))

	_, err = clientset.CoreV1().ConfigMaps(c.CMNamespace).Patch(context.Background(), c.CMName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return createConfigMap(clientset, c, permissions)
		}
	}
	return nil
}

func createConfigMap(clientset *kubernetes.Clientset, c Config, permissions []byte) error {
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.CMName,
			Namespace: c.CMNamespace,
		},
		Immutable: nil,
		Data: map[string]string{
			"labels.yaml": string(permissions),
		},
		BinaryData: nil,
	}
	_, err := clientset.CoreV1().ConfigMaps(c.CMNamespace).Create(context.Background(), &cm, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func MapsEqual(m1, m2 map[string]map[string]bool) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, val1 := range m1 {
		val2, ok := m2[k]
		if !ok {
			return false
		}

		if len(val1) != len(val2) {
			return false
		}

		for kk, vv1 := range val1 {
			vv2, exists := val2[kk]
			if !exists || vv1 != vv2 {
				return false
			}
		}
	}
	return true
}
