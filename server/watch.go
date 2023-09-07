package server

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rs/zerolog/log"

	v1r "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func watchResources(resource ResourceInterface, wrapper *ResourceListWrapper, s chan struct{}) {
	for {
		err := error(nil)
		list, err := resource.List(metav1.ListOptions{})
		if err != nil {
			log.Fatal().Err(err).Msgf("Error listing resources")
			return
		}
		wrapper.Update(list) // Here, update the wrapper's list
		meta, ok := list.(metav1.ListInterface)
		if !ok {
			log.Fatal().Msg("Expected metav1.ListInterface")
		}

		resourceVersion := meta.GetResourceVersion()

		opts := metav1.ListOptions{ResourceVersion: resourceVersion}
		watcher, err := resource.Watch(opts)
		if err != nil {
			log.Fatal().Err(err).Msgf("Error watching resources")
			return
		}

		for event := range watcher.ResultChan() {
			handleEvent(event, wrapper.List)
			s <- struct{}{}
		}
		log.Error().Msg("Error watching resources, reconnecting...")
	}
}

// handleEvent will process watch events and modify the provided resourceList accordingly.
// This can be extended to handle different types of resources.
func handleEvent(event watch.Event, resourceList runtime.Object) {
	switch obj := resourceList.(type) {
	case *v1r.ClusterRoleBindingList:
		crb, ok := event.Object.(*v1r.ClusterRoleBinding)
		if !ok {
			log.Error().Msgf("Unexpected type %T", event.Object)
			return
		}
		switch event.Type {
		case watch.Added:
			log.Debug().Msgf("%T added: %s", crb, crb.Name)
			obj.Items = append(obj.Items, *crb)
		case watch.Modified:
			log.Debug().Msgf("%T modified: %s", crb, crb.Name)
			for i, item := range obj.Items {
				if item.UID == crb.UID {
					obj.Items[i] = *crb
					break
				}
			}
		case watch.Deleted:
			log.Debug().Msgf("%T deleted: %s", crb, crb.Name)
			for i, item := range obj.Items {
				if item.UID == crb.UID {
					obj.Items = append(obj.Items[:i], obj.Items[i+1:]...)
					break
				}
			}
		case watch.Error:
			log.Error().Msg("Error watching ClusterRoleBinding, reconnecting...")
			time.Sleep(5 * time.Second)
			handleEvent(event, resourceList)
		default:
			log.Error().Msgf("Unexpected event type %s", event.Type)
		}
	case *v1r.RoleBindingList:
		rb, ok := event.Object.(*v1r.RoleBinding)
		if !ok {
			log.Error().Msgf("Unexpected type %T", event.Object)
			return
		}
		switch event.Type {
		case watch.Added:
			log.Debug().Msgf("%T added: %s", rb, rb.Name)
			obj.Items = append(obj.Items, *rb)
		case watch.Modified:
			log.Debug().Msgf("%T modified: %s", rb, rb.Name)
			for i, item := range obj.Items {
				if item.UID == rb.UID {
					obj.Items[i] = *rb
					break
				}
			}
		case watch.Deleted:
			log.Debug().Msgf("%T deleted: %s", rb, rb.Name)
			for i, item := range obj.Items {
				if item.UID == rb.UID {
					obj.Items = append(obj.Items[:i], obj.Items[i+1:]...)
					break
				}
			}
		case watch.Error:
			log.Debug().Msg("Error watching RoleBinding, reconnecting...")
			time.Sleep(5 * time.Second)
			handleEvent(event, resourceList)
		default:
			log.Error().Msgf("Unexpected event type %s", event.Type)
		}
	case *v1r.RoleList:
		r, ok := event.Object.(*v1r.Role)
		if !ok {
			log.Error().Msgf("Unexpected type %T", event.Object)
			return
		}
		switch event.Type {
		case watch.Added:
			log.Debug().Msgf("%T added: %s", r, r.Name)
			obj.Items = append(obj.Items, *r)
		case watch.Modified:
			log.Debug().Msgf("%T modified: %s", r, r.Name)
			for i, item := range obj.Items {
				if item.UID == r.UID {
					obj.Items[i] = *r
					break
				}
			}
		case watch.Deleted:
			log.Debug().Msgf("%T deleted: %s", r, r.Name)
			for i, item := range obj.Items {
				if item.UID == r.UID {
					obj.Items = append(obj.Items[:i], obj.Items[i+1:]...)
					break
				}
			}
		case watch.Error:
			log.Error().Msg("Error watching Role, reconnecting...")
			time.Sleep(5 * time.Second)
			handleEvent(event, resourceList)
		default:
			log.Error().Msgf("Unexpected event type %s", event.Type)
		}
	case *v1r.ClusterRoleList:
		cr, ok := event.Object.(*v1r.ClusterRole)
		if !ok {
			log.Error().Msgf("Unexpected type %T", event.Object)
			return
		}
		switch event.Type {
		case watch.Added:
			log.Debug().Msgf("%T added: %s", cr, cr.Name)
			obj.Items = append(obj.Items, *cr)
		case watch.Modified:
			log.Debug().Msgf("%T modified: %s", cr, cr.Name)
			for i, item := range obj.Items {
				if item.UID == cr.UID {
					obj.Items[i] = *cr
					break
				}
			}
		case watch.Deleted:
			log.Debug().Msgf("%T deleted: %s", cr, cr.Name)
			for i, item := range obj.Items {
				if item.UID == cr.UID {
					obj.Items = append(obj.Items[:i], obj.Items[i+1:]...)
					break
				}
			}
		case watch.Error:
			log.Error().Msg("Error watching ClusterRole, reconnecting...")
			time.Sleep(5 * time.Second)
			handleEvent(event, resourceList)
		default:
			log.Error().Msgf("Unexpected event type %s", event.Type)
		}
	default:
		log.Error().Msgf("Unexpected type %T", resourceList)
	}
}
