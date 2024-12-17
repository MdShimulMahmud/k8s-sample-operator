package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	webappv1 "my.domain/guestbook/api/v1"
)

const GuestbookFinalizer = "my.domain/dev-protection"

func (r *GuestbookReconciler) DeploymentCreation(ctx context.Context, req ctrl.Request) error {
	log := log.FromContext(ctx)
	guestbook := &webappv1.Guestbook{}
	deployment := &appsv1.Deployment{}

	err := r.Get(ctx, req.NamespacedName, guestbook)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "Guestbook not found")
			return err
		}
		log.Error(err, "Failed to get Guestbook resource")
		return err
	}

	if !guestbook.ObjectMeta.DeletionTimestamp.IsZero() {
		guestbook.Finalizers = nil
		if err := r.Update(ctx, guestbook); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}

	}

	err = r.Get(ctx, req.NamespacedName, deployment)
	if err == nil {
		log.Info("Deployment exists, updating", "Deployment", deployment.Name)
		if err = r.Update(ctx, deployment); err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment", deployment.Name)
			return err
		}
		log.Info("Updated Deployment", "Deployment", deployment.Name)
	} else if errors.IsNotFound(err) {
		log.Info("Creating new Deployment", "Deployment", guestbook.Name)

		deployment = &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      guestbook.Name,
				Namespace: req.Namespace,
				Finalizers: []string{
					GuestbookFinalizer,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &guestbook.Spec.Replicas,
				Selector: &v1.LabelSelector{
					MatchLabels: map[string]string{"app": "guestbook"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{"app": "guestbook"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "guestbook-container",
								Image: guestbook.Spec.ImageName,
								Ports: []corev1.ContainerPort{{ContainerPort: 80}},
							},
						},
					},
				},
			},
		}
		if err = r.Create(ctx, deployment); err != nil {
			log.Error(err, "Failed to create Deployment", "Deployment", guestbook.Name)
			return err
		}
		log.Info("Created new Deployment", "Deployment", guestbook.Name)
	} else {
		log.Error(err, "Failed to get Deployment")
		return err
	}

	return nil
}
