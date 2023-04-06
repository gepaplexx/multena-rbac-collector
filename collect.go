package main

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/context"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func collectAll() (map[string][]string, map[string][]string, error) {
	bar := progressbar.NewOptions(3,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(45),
		progressbar.OptionSetDescription("[cyan][1/3][reset] Collecting namespaces, users and groups ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>🐆[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	bar.Describe("[cyan][1/3][reset] Now Collecting namespaces ...")
	namespaceList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	var namespaces []string
	for _, namespace := range namespaceList.Items {
		if !strings.HasPrefix(namespace.ObjectMeta.Name, "openshift") && !strings.HasPrefix(namespace.ObjectMeta.Name, "kube") {
			namespaces = append(namespaces, namespace.ObjectMeta.Name)
		}
	}
	err = bar.Add(1)
	if err != nil {
		return nil, nil, err
	}

	bar.Describe("[cyan][1/3][reset] Now Collecting groups ...")
	groupList, err := userClientset.UserV1().Groups().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	// Extract the names of the groups and return as a slice of strings
	var groups []string
	for _, group := range groupList.Items {
		groups = append(groups, group.ObjectMeta.Name)
	}
	err = bar.Add(1)
	if err != nil {
		return nil, nil, err
	}

	bar.Describe("[cyan][1/3][reset] Now Collecting Users ...")
	userList, err := userClientset.UserV1().Users().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	// Extract the names of the users and return as a slice of strings
	var users []string
	for _, user := range userList.Items {
		users = append(users, user.ObjectMeta.Name)
	}
	err = bar.Add(1)
	if err != nil {
		return nil, nil, err
	}

	bar = progressbar.NewOptions(len(users)*len(namespaces),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(45),
		progressbar.OptionSetDescription("[cyan][2/3][reset] Checking SAR for users ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>🐆[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	userAllowedNamespaces := make(map[string][]string)
	for _, user := range users {
		bar.Describe(fmt.Sprintf("[cyan][2/3][reset] Checking access for user %s  ...", user))
		for _, ns := range namespaces {
			allowed, err := checkAccessForUserOrGroup(user, ns, "get", "pods")
			if err != nil {
				return nil, nil, err
			}
			if allowed {
				userAllowedNamespaces[user] = append(userAllowedNamespaces[user], ns)
			}
			err = bar.Add(1)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	bar = progressbar.NewOptions(len(groups)*len(namespaces),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionSetWidth(45),
		progressbar.OptionSetDescription("[cyan][3/3][reset] Checking SAR for groups ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>🐆[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	groupAllowedNamespaces := make(map[string][]string)
	for _, group := range groups {
		bar.Describe(fmt.Sprintf("[cyan][2/3][reset] Checking access for group %s ...", group))
		for _, ns := range namespaces {
			allowed, err := checkAccessForUserOrGroup(group, ns, "get", "pods")
			if err != nil {
				return nil, nil, err
			}
			if allowed {
				groupAllowedNamespaces[group] = append(groupAllowedNamespaces[group], ns)
			}
			err = bar.Add(1)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return userAllowedNamespaces, groupAllowedNamespaces, nil
}

func checkAccessForUserOrGroup(userOrGroup string, namespace string, verb string, resource string) (bool, error) {
	// Create a new SubjectAccessReview
	sar := &v1.SubjectAccessReview{
		Spec: v1.SubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Verb:      verb,
				Group:     "",
				Resource:  resource,
				Namespace: namespace,
			},
			User:   userOrGroup,
			Groups: []string{userOrGroup},
		},
	}

	// Submit the SubjectAccessReview
	response, err := clientset.AuthorizationV1().SubjectAccessReviews().Create(context.Background(), sar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}

	// Check the result of the SubjectAccessReview
	return response.Status.Allowed, nil
}
