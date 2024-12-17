package controller

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
		return client.IgnoreNotFound(err) // Ignore not-found errors
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

	// Check if the Service already exists
	foundService := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: guestbook.Name, Namespace: guestbook.Namespace}, foundService)
	if err != nil && client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to get Service")
		return err
	}

	if err != nil && client.IgnoreNotFound(err) == nil {
		// Service does not exist, create a new one
		log.Info("Creating a new Service", "Service.Name", service.Name)
		if err := controllerutil.SetOwnerReference(guestbook, service, r.Scheme); err != nil {
			log.Error(err, "Failed to set owner reference for Service")
			return err
		}
		if err := r.Create(ctx, service); err != nil {
			log.Error(err, "Failed to create Service")
			return err
		}
		log.Info("Created new Service", "Service.Name", service.Name)
		return nil
	}

	// Service already exists - no need to update unless there are changes
	log.Info("Service already exists", "Service.Name", foundService.Name)
	return nil
}
