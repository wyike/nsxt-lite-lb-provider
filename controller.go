package main

import (
	"context"

	discoveryv1 "k8s.io/api/discovery/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EndpointSliceReconciler watches EndpointSlice events.
type EndpointSliceReconciler struct {
	client.Client
	NsxClient *NSXClient
}

// Reconcile function will be called when an EndpointSlice is created, updated, or deleted.
func (r *EndpointSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the EndpointSlice
	var slice discoveryv1.EndpointSlice
	if err := r.Get(ctx, req.NamespacedName, &slice); err != nil {
		log.WithName("endpointSliceReconciler").Error(err, "unable to fetch kubernetes EndpointSlice")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	endpoints := []string{}
	for _, endpoint := range slice.Endpoints {
		if len(endpoint.Addresses) > 0 {
			endpoints = append(endpoints, endpoint.Addresses...)
		}
	}
	log.WithName("endpointSliceReconciler").Info("processing EndpointSlice", "name", slice.Name, "endpoints", endpoints)
	return ctrl.Result{}, r.syncRoutes(endpoints)
}

func (r *EndpointSliceReconciler) syncRoutes(endpoints []string) error {
	return r.NsxClient.PatchStaticRoute(endpoints)
}
