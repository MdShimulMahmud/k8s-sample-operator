package controller

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	webappv1 "my.domain/guestbook/api/v1"
)

func (r *GuestbookReconciler) ServiceCreation(ctx context.Context, req ctrl.Request) error {
	log := log.FromContext(ctx)

	// Fetch the Guestbook instance
	guestbook := &webappv1.Guestbook{}
	if err := r.Get(ctx, req.NamespacedName, guestbook); err != nil {
		log.Error(err, "Failed to get Guestbook")
		return err
	}

	// Define the Service
	service := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      guestbook.Name,
			Namespace: guestbook.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "guestbook"},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
	err := r.Get(ctx, client.ObjectKey{Name: guestbook.Name, Namespace: guestbook.Namespace}, service)
	if err != nil && client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to get Service, creating new Service")
		if err := r.Create(ctx, service); err != nil {
			log.Error(err, "Failed to create Service")
			return err
		}
		log.Info("Created new Service", "name", guestbook.Name)
		return nil
	}
	log.Info("Updating Service", "name", guestbook.Name)
	if err := r.Update(ctx, service); err != nil {
		log.Error(err, "Failed to update Service")
		return err
	}
	log.Info("Updated Service", "name", guestbook.Name)

	return nil
}
