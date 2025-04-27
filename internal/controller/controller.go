package controller

import (
	"context"
	"fmt"

	v1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type IngressController struct {
	clientset *kubernetes.Clientset
	tunnelID  string
}

func NewIngressController(clientset *kubernetes.Clientset, tunnelID string) *IngressController {
	return &IngressController{
		clientset: clientset,
		tunnelID:  tunnelID,
	}
}

func (ic *IngressController) ProcessIngress(ctx context.Context, ing *v1.Ingress) error {
	if ing.Spec.IngressClassName == nil || *ing.Spec.IngressClassName != "cloudflare" {
		return nil
	}

	if len(ing.Status.LoadBalancer.Ingress) > 0 &&
		ing.Status.LoadBalancer.Ingress[0].Hostname == ic.tunnelID+".cfargotunnel.com" {
		return nil
	}

	return ic.updateIngressStatus(ctx, ing)
}

func (ic *IngressController) updateIngressStatus(ctx context.Context, ing *v1.Ingress) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		updatedIngress, err := ic.clientset.NetworkingV1().Ingresses(ing.Namespace).Get(ctx, ing.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		updatedIngress.Status.LoadBalancer.Ingress = []v1.IngressLoadBalancerIngress{
			{Hostname: ic.tunnelID + ".cfargotunnel.com"},
		}

		_, err = ic.clientset.NetworkingV1().Ingresses(ing.Namespace).UpdateStatus(ctx, updatedIngress, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Failed to update status for %s/%s: %v\n", ing.Namespace, ing.Name, err)
			return err
		}

		fmt.Printf("Updated status for %s/%s\n", ing.Namespace, ing.Name)
		return nil
	})
}
