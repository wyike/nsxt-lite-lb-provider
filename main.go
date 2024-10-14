package main

import (
	"flag"
	"os"

	discoveryv1 "k8s.io/api/discovery/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	log = ctrl.Log.WithName("nsxt-lite-lb-provider")
)

const (
	Kubernetes = "kubernetes"
	Namespace  = "default"
)

func main() {
	var nsxHost, nsxUser, nsxPassword, clusterVIP, clusterRouter string
	var setVIPEnabled bool
	flag.StringVar(&nsxHost, "nsxhost", "https://nsxhost", "nsx host name, please provide with scheme like: https://nsxhost")
	flag.StringVar(&nsxUser, "nsxuser", "admin", "nsx user name")
	flag.StringVar(&nsxPassword, "nsxpassword", "prettycoco", "nsx password")
	flag.StringVar(&clusterVIP, "clustervip", "10.10.10.1", "cluster controlplane VIP, that is defined by user")
	flag.StringVar(&clusterRouter, "clusterrouterid", "pks-a898fb55-f7dc-4e8e-b4ea-c9eb1d08a46a-cluster-router", "tier1 policy id, that controller plane vms are connected to")
	flag.BoolVar(&setVIPEnabled, "setvipenabled", true, "vip configuration on control plane is enabled or not. It requires privilege to run system command on VM")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	if setVIPEnabled {
		if err := SetLoopbackIP(clusterVIP); err != nil {
			log.Error(err, "failed to set VIP on current control plane vm")
			os.Exit(1)
		}
	}

	// Use controller-runtime's built-in GetConfig to load the kubeconfig
	// Use --kubeconfig running out of cluster, controller-runtime by default registers this flag
	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}

	// Set up a manager
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:          true,
		LeaderElectionID:        "nsxt-lite-lb-controller-leader-election",
		LeaderElectionNamespace: Namespace,
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				Namespace: {},
			},
		},
	})
	if err != nil {
		log.Error(err, "unable to create controller manager")
		os.Exit(1)
	}

	// Add the EndpointSliceReconciler to the manager
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.EndpointSlice{}).
		WithEventFilter(SpecificEndpointSliceFilter(Kubernetes, Namespace)).
		Complete(&EndpointSliceReconciler{
			Client:    mgr.GetClient(),
			NsxClient: NewNSXClient(nsxHost, nsxUser, nsxPassword, clusterVIP, clusterRouter),
		}); err != nil {
		log.Error(err, "unable to create controller")
		os.Exit(1)
	}

	// Start the manager
	ctrl.Log.Info("starting nsxt lite lb controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "cloud not start manager")
		os.Exit(1)
	}
}

func SpecificEndpointSliceFilter(sliceName, namespace string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetName() == sliceName && e.Object.GetNamespace() == namespace
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.GetName() == sliceName && e.ObjectNew.GetNamespace() == namespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object.GetName() == sliceName && e.Object.GetNamespace() == namespace
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return e.Object.GetName() == sliceName && e.Object.GetNamespace() == namespace
		},
	}
}
