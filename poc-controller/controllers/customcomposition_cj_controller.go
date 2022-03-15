/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/api/batch/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pocv1alpha1 "sergenyalcin.io/poc/api/v1alpha1"
)

// CustomCompositionCjReconciler reconciles a CustomComposition object
type CustomCompositionCjReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=poc.sergenyalcin.io,resources=customcompositions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=poc.sergenyalcin.io,resources=customcompositions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=poc.sergenyalcin.io,resources=customcompositions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomComposition object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *CustomCompositionCjReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("customcompositioncj", req.NamespacedName)

	log.Info("reconciliation started")

	var customComposition pocv1alpha1.CustomComposition
	if err := r.Get(ctx, req.NamespacedName, &customComposition); err != nil {
		log.Error(err, "unable to fetch CustomComposition")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizer := "customcomposition.poc.crossplane.io/finalizer"

	if customComposition.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&customComposition, finalizer) {
			controllerutil.AddFinalizer(&customComposition, finalizer)
			if err := r.Update(ctx, &customComposition); err != nil {
				log.Error(err, "unable to add finalizer")
				return ctrl.Result{}, err
			}
			log.Info("finalizer successfully added")
		}
	} else {
		if controllerutil.ContainsFinalizer(&customComposition, finalizer) {
			if err := r.Delete(ctx, &v1beta1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      customComposition.Name,
					Namespace: customComposition.Namespace,
				},
			}); err != nil {
				log.Error(err, "unable to delete cronjob")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			log.Info("cronjob successfully deleted")
			controllerutil.RemoveFinalizer(&customComposition, finalizer)
			if err := r.Update(ctx, &customComposition); err != nil {
				log.Error(err, "unable to delete finalizer")
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	cj := &v1beta1.CronJob{}
	if err := r.Get(ctx, types.NamespacedName{Name: customComposition.Name, Namespace: customComposition.Namespace}, cj); err != nil {
		if errors.IsNotFound(err) {
			cj = constructCronJob(&customComposition)
			if err := r.Create(ctx, cj); err != nil {
				log.Error(err, "unable to create the cronjob")
				return ctrl.Result{}, err
			}
			log.Info("cronjob was successfully created")
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		} else {
			log.Error(err, "unable to fetch cronjob")
		}
	}

	cj = constructCronJob(&customComposition)
	if err := r.Update(ctx, cj); err != nil {
		log.Error(err, "unable to update the cronjob")
		return ctrl.Result{}, err
	}

	customComposition.Status.Conditions = []pocv1alpha1.StatusCondition{{Message: cj.Status.String()}}

	if err := r.Status().Update(ctx, &customComposition); err != nil {
		log.Error(err, "unable to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func constructCronJob(customComposition *pocv1alpha1.CustomComposition) *v1beta1.CronJob {
	initContainerVms := []corev1.VolumeMount{
		{
			Name:      "storage",
			MountPath: "/data/",
		},
	}

	cj := &v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customComposition.Name,
			Namespace: customComposition.Namespace,
		},
		Spec: v1beta1.CronJobSpec{
			Schedule: "*/3 * * * *",
		},
	}

	cj.Spec.JobTemplate.Spec.Template.Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:         "debug-resources",
				Image:        baseImage,
				Command:      []string{"sh", "-c", "sleep 60"},
				VolumeMounts: initContainerVms,
			},
		},
		InitContainers: []corev1.Container{
			{
				Name:         "prep-resources",
				Image:        baseImage,
				Command:      []string{"sh", "-c", fmt.Sprintf(`echo "%s" >> /data/resource.yaml`, customComposition.Spec.Resource)},
				VolumeMounts: initContainerVms,
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "storage",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
		RestartPolicy: corev1.RestartPolicyNever,
	}

	for _, f := range customComposition.Spec.Functions {
		cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers = append(cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers, corev1.Container{
			Name:  f.Title,
			Image: f.Image,
			Command: []string{"sh", "-c", fmt.Sprintf("kpt fn eval /data --exec /function -- %s",
				f.Args)},
			VolumeMounts: initContainerVms,
		})
	}

	return cj
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomCompositionCjReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pocv1alpha1.CustomComposition{}).
		Complete(r)
}
