package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gepaplexx/multena-rbac-collector/collector"
	"github.com/gepaplexx/multena-rbac-collector/util"
	v1r "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
)

func Serve(clientset *kubernetes.Clientset, port int, config util.Config) {
	signal := make(chan struct{}, 100000)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Ok")
	})

	http.HandleFunc("/invoke", func(w http.ResponseWriter, r *http.Request) {
		signal <- struct{}{}
		_, err := fmt.Fprintf(w, "Invoked")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	go Watch(clientset, signal, config)
	// Start the server
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		return
	}
	log.Error().Msg("This should not be printed")
	select {}
}

func Watch(clientset *kubernetes.Clientset, signal chan struct{}, config util.Config) {
	crbList := ResourceListWrapper{&v1r.ClusterRoleBindingList{}}
	go watchResources(&ClusterRoleBindingAdapter{client: clientset}, &crbList, signal)

	rbList := ResourceListWrapper{&v1r.RoleBindingList{}}
	go watchResources(&RoleBindingAdapter{client: clientset}, &rbList, signal)

	roleList := ResourceListWrapper{&v1r.RoleList{}}
	go watchResources(&RoleAdapter{client: clientset}, &roleList, signal)

	crList := ResourceListWrapper{&v1r.ClusterRoleList{}}
	go watchResources(&ClusterRoleAdapter{client: clientset}, &crList, signal)

	time.AfterFunc(2*time.Second, func() { signal <- struct{}{} })

	currentPermission := make(map[string]map[string]bool, 1000)

	for range signal {
		log.Debug().Msg("received signal")
		roles, clusterRoles := collector.GetRoles(*roleList.List.(*v1r.RoleList), *crList.List.(*v1r.ClusterRoleList))
		permissions := collector.Collect(roles, clusterRoles, rbList.List.(*v1r.RoleBindingList), crbList.List.(*v1r.ClusterRoleBindingList))
		if !util.MapsEqual(currentPermission, permissions) {
			currentPermission = permissions
			err := util.WriteConfigmap(clientset, permissions, config)
			if err != nil {
				log.Fatal().Err(err).Msg("Error writing configmap")
				return
			}
			log.Info().Msg("Configmap updated")
		}
		for range signal {
			// this lets one signal in the channel for the rare occurrence that while draining the channel a new signal is received
			if len(signal) > 1 {
				<-signal
			} else {
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
}
