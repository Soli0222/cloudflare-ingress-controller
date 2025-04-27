package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	ingresscontroller "github.com/Soli0222/cloudflare-ingress-controller/internal/controller" // Renamed import

	v1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to the kubeconfig file (optional)")
	tunnelIDFlag := flag.String("tunnel-id", "", "Cloudflare tunnel ID (for cfargotunnel.com). Can also be set via TUNNEL_ID environment variable.")
	flag.Parse()

	// Get Tunnel ID: prioritize environment variable, then flag
	tunnelID := os.Getenv("TUNNEL_ID")
	if tunnelID == "" {
		tunnelID = *tunnelIDFlag
	}

	// Validate Tunnel ID
	if tunnelID == "" {
		fmt.Println("Error: Tunnel ID is required. Set it via the --tunnel-id flag or the TUNNEL_ID environment variable.")
		os.Exit(1)
	}

	// Determine kubeconfig path: use flag value or default behavior if empty
	kubeconfigPath := *kubeconfig
	// If the kubeconfig flag was not provided, BuildConfigFromFlags will try in-cluster config or default paths
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)

	// Create an instance of the IngressController
	ic := ingresscontroller.NewIngressController(clientset, tunnelID) // Use the determined tunnelID

	ingressInformer := factory.Networking().V1().Ingresses().Informer()
	ctx := context.Background() // Use context.Background() instead of context.TODO()

	ingressInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ing, ok := obj.(*v1.Ingress)
			if !ok {
				fmt.Println("Error: received non-Ingress object in AddFunc")
				return
			}
			// Call ProcessIngress on the controller instance
			if err := ic.ProcessIngress(ctx, ing); err != nil {
				fmt.Printf("Error processing added Ingress %s/%s: %v\n", ing.Namespace, ing.Name, err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ing, ok := newObj.(*v1.Ingress)
			if !ok {
				fmt.Println("Error: received non-Ingress object in UpdateFunc")
				return
			}
			// Call ProcessIngress on the controller instance
			if err := ic.ProcessIngress(ctx, ing); err != nil {
				fmt.Printf("Error processing updated Ingress %s/%s: %v\n", ing.Namespace, ing.Name, err)
			}
		},
		// DeleteFunc is usually not needed for this controller's logic,
		// as we only care about updating existing Ingresses.
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	version := os.Getenv("VERSION")
	fmt.Printf("Starting Cloudflare Ingress Controller %s with Tunnel ID: %s\n", version, tunnelID)
	fmt.Println("Watching Ingresses with ingressClassName=cloudflare...")

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, ingressInformer.HasSynced) {
		fmt.Println("Error: Timed out waiting for caches to sync")
		os.Exit(1)
	}

	fmt.Println("Cache synced, controller is running.")
	<-stopCh
	fmt.Println("Shutting down controller.")
}
