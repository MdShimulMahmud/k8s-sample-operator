package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "my.domain/guestbook/api/v1"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const GuestbookFinalizer = "my.domain/dev-protection"

func (r *GuestbookReconciler) AddDeploymentFinalizer(ctx context.Context, req ctrl.Request) error {
	log := log.FromContext(ctx)
	guestbook := &webappv1.Guestbook{}

	if err := r.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, guestbook); err != nil {
		return err
	}

	if guestbook.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(guestbook, GuestbookFinalizer) {
			log.Info("Adding finalizer to Guestbook", "Guestbook", guestbook.Name)
			controllerutil.AddFinalizer(guestbook, GuestbookFinalizer)
			if err := r.Update(ctx, guestbook); err != nil {
				log.Error(err, "Failed to update Guestbook after adding finalizer", "Guestbook", guestbook.Name)
				return err
			}
		}
	} else {
		// deletion request
		if controllerutil.ContainsFinalizer(guestbook, GuestbookFinalizer) {
			controllerutil.RemoveFinalizer(guestbook, GuestbookFinalizer)
			log.Info("Removing finalizer from Guestbook", "Guestbook", guestbook.Name)
			if err := r.Update(ctx, guestbook); err != nil {
				log.Error(err, "Failed to update Guestbook", "Guestbook", guestbook.Name)
				return err
			}
		}
	}

	return nil
}
