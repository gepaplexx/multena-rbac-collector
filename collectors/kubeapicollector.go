package collectors

import (
	"context"
	"fmt"
	"github.com/openshift/client-go/user/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/authorization/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"rbac-collector/config"
	"strings"
	"sync"
	"time"
)

type KubeApiCollector struct {
	kubeconfig    *rest.Config
	clientset     *kubernetes.Clientset
	userClientset *versioned.Clientset
	config        *config.Configuration
}

type Subjects struct {
	namespaces, users, groups []string
}

var ignoredNamespacePrefixes = []string{
	"kube-",
	"openshift",
	"syn",
	"vshn",
	"appuio",
	"cilium",
}

func NewKubeApiCollector(c config.Configuration) (*KubeApiCollector, error) {
	kubeconfig := buildKubeconfig(c)
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Errorf("error creating clientset: %v", err)
		return nil, err
	}

	userClientset, err := versioned.NewForConfig(kubeconfig)
	if err != nil {
		log.Errorf("error creating user clientset: %v", err)
		return nil, err
	}

	return &KubeApiCollector{
		kubeconfig:    kubeconfig,
		clientset:     clientset,
		userClientset: userClientset,
		config:        &c,
	}, nil
}

func buildKubeconfig(c config.Configuration) *rest.Config {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err == nil {
		return kubeconfig
	}
	log.Info("Couldn't build kubeconfig from parameters, trying in-cluster config")
	kubeconfig, err = rest.InClusterConfig()
	if err != nil {
		log.Errorf("error creating in-cluster config: %v", err)
		return nil
	}
	return kubeconfig
}

func (c *KubeApiCollector) Collect() {
	start := time.Now()
	log.Info("KubeApiCollector starting, this is going to take a while ...")
	wg := sync.WaitGroup{}
	groupAllowList := make(map[string][]string)
	userAllowList := make(map[string][]string)

	subjects := Subjects{
		namespaces: []string{},
		users:      []string{},
		groups:     []string{},
	}

	wg.Add(1)
	log.Info("Collecting information about users, groups and namespaces")
	go collectSubjects(&subjects, &wg, c)
	wg.Wait()

	wg.Add(2)
	log.Info("Collecting rbac permission for users and groups")
	go c.collectFromApi(subjects.groups, subjects.namespaces, groupAllowList, &wg)
	go c.collectFromApi(subjects.users, subjects.namespaces, userAllowList, &wg)
	wg.Wait()
	log.Info("Finished collecting rbac permissions")

	err := createOrUpdateConfigMap(c, groupAllowList, userAllowList)
	if err != nil {
		log.Errorf("error creating or updating configmap: %e", err)
	}
	log.Infof("Collection took %s", time.Since(start))

}

func collectSubjects(s *Subjects, wg *sync.WaitGroup, c *KubeApiCollector) {
	defer wg.Done()
	wg.Add(3)
	go collectNamespaces(s, c, wg)
	go collectUsers(s, c, wg)
	go collectGroups(s, c, wg)
}

func (c *KubeApiCollector) collectFromApi(subjects []string, namespaces []string, collection map[string][]string, wg *sync.WaitGroup) {
	defer wg.Done()

	processSubject := func(s string, ns []string) {
		for _, namespace := range ns {
			allowed, err := c.checkAccess(s, namespace)
			if err != nil {
				log.Errorf("error while checking access: %v", err)
				return
			}
			if allowed {
				// Mutex protects shared map from concurrent write access
				collection[s] = append(collection[s], namespace)
			}
		}
	}

	for _, subject := range subjects {
		processSubject(subject, namespaces)
	}
}

func (c *KubeApiCollector) checkAccess(subject string, ns string) (bool, error) {
	sar := createSubjectAccessReview(subject, ns)
	log.Debugf("Checking access for %s in namespace %s", subject, ns)
	response, err := c.clientset.AuthorizationV1().SubjectAccessReviews().Create(context.Background(), sar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("error while checking access: %v", err)
	}
	return response.Status.Allowed, nil
}

func createSubjectAccessReview(subject string, ns string) *v1.SubjectAccessReview {
	return &v1.SubjectAccessReview{
		Spec: v1.SubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Verb:      "get",
				Group:     "",
				Resource:  "pods",
				Namespace: ns,
			},
			User:   subject,
			Groups: []string{subject},
		},
	}
}

func collectNamespaces(s *Subjects, c *KubeApiCollector, wg *sync.WaitGroup) {
	defer wg.Done()

	resourceList, err := c.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Error while collecting namespaces: %v", err)
		return
	}

	for _, resource := range resourceList.Items {
		s.namespaces = appendResourceNames(resource.Name, ignoredNamespacePrefixes, s.namespaces)
	}
}

func collectUsers(s *Subjects, c *KubeApiCollector, wg *sync.WaitGroup) {
	defer wg.Done()

	resourceList, err := c.userClientset.UserV1().Users().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Error while collecting users: %v", err)
		return
	}

	for _, resource := range resourceList.Items {
		s.users = append(s.users, resource.Name)
	}
}

func collectGroups(s *Subjects, c *KubeApiCollector, wg *sync.WaitGroup) {
	defer wg.Done()

	resourceList, err := c.userClientset.UserV1().Groups().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Error while collecting groups: %v", err)
		return
	}

	for _, resource := range resourceList.Items {
		s.groups = append(s.groups, resource.Name)
	}
}

func appendResourceNames(resourceName string, ignoredPrefixes []string, resourceSlice []string) []string {
	if !ContainsAnySequence(resourceName, ignoredPrefixes) {
		return append(resourceSlice, resourceName)
	}

	return resourceSlice
}

func createOrUpdateConfigMap(c *KubeApiCollector, groups map[string][]string, users map[string][]string) error {
	tenants := map[string]map[string][]string{"users": users, "groups": groups}
	permissions, err := yaml.Marshal(tenants)
	if err != nil {
		log.Errorf("Failed to marshal yaml: %v", err)
		return err
	}

	cm := &v1core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.config.ConfigMapName,
			Namespace: c.config.Namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			"labels.yaml": string(permissions),
		},
		BinaryData: nil,
	}

	_, err = c.clientset.CoreV1().ConfigMaps(c.config.Namespace).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err := c.clientset.CoreV1().ConfigMaps(c.config.Namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
			if err != nil {
				log.Errorf("Failed to update configmap: %v", err)
				return err
			}
		} else {
			log.Errorf("Failed to create configmap: %v", err)
			return err
		}
	}

	return nil
}

func ContainsAnySequence(s string, sequences []string) bool {
	for _, seq := range sequences {
		if strings.HasPrefix(s, seq) {
			return true
		}
	}
	return false
}
