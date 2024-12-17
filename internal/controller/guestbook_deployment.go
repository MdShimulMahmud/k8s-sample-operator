package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	webappv1 "my.domain/guestbook/api/v1"
)

func (r *GuestbookReconciler) DeploymentCreation(ctx context.Context, req ctrl.Request) error {
	log := log.FromContext(ctx)
	guestbook := &webappv1.Guestbook{}
	deployment := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, guestbook)

	if err != nil {
		return err
	}

	err = r.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, deployment)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating new Deployment", "Deployment", guestbook.Name)

		deployment = &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      guestbook.Name,
				Namespace: req.Namespace,
				OwnerReferences: []v1.OwnerReference{

					{
						APIVersion: guestbook.APIVersion,
						Kind:       guestbook.Kind,
						Name:       guestbook.Name,
						UID:        guestbook.UID,
						Controller: ptr.To(true),
					},
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
			log.Error(err, "Failed to create Deployment", "Deployment:::", err.Error())
			return err
		}
		log.Info("Created new Deployment", "Deployment", guestbook.Name)
	} else {
		log.Info("Deployment already exists")
		if guestbook.Spec.Replicas != *deployment.Spec.Replicas {
			deployment.Spec.Replicas = &guestbook.Spec.Replicas
		}

		if err := r.Update(ctx, deployment); err != nil {
			log.Error(err, "Failed to update image!")
			return err
		}

		return nil
	}

	return nil
}
