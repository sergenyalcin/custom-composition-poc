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

const (
	baseImage = "sergenyalcin10/custom-composition-poc:3.0-alpha"
)

// CustomCompositionReconciler reconciles a CustomComposition object
type CustomCompositionReconciler struct {
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
func (r *CustomCompositionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("customcomposition", req.NamespacedName)

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
			if err := r.Delete(ctx, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      customComposition.Name,
					Namespace: customComposition.Namespace,
				},
			}); err != nil {
				log.Error(err, "unable to delete pod")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			log.Info("pod successfully deleted")
			controllerutil.RemoveFinalizer(&customComposition, finalizer)
			if err := r.Update(ctx, &customComposition); err != nil {
				log.Error(err, "unable to delete finalizer")
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	pod := &corev1.Pod{}
	if err := r.Get(ctx, types.NamespacedName{Name: customComposition.Name, Namespace: customComposition.Namespace}, pod); err != nil {
		if errors.IsNotFound(err) {
			pod = constructPod(&customComposition)
			if err := r.Create(ctx, pod); err != nil {
				log.Error(err, "unable to create the pod")
				return ctrl.Result{}, err
			}
			log.Info("pod was successfully created")
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		} else {
			log.Error(err, "unable to fetch pod")
		}
	}

	customComposition.Status.Conditions = nil
	for _, cs := range pod.Status.InitContainerStatuses {
		condition := pocv1alpha1.StatusCondition{
			Operation: cs.Name,
			Completed: &cs.Ready,
		}

		if cs.State.Waiting != nil {
			condition.Message = fmt.Sprintf("%s %s", cs.State.Waiting.Reason, cs.State.Waiting.Message)
		} else if cs.State.Running != nil {
			condition.Message = fmt.Sprintf("%s", cs.State.Running.StartedAt)
		} else if cs.State.Terminated != nil {
			condition.Message = fmt.Sprintf("%s", cs.State.Terminated.Reason)
		}

		customComposition.Status.Conditions = append(customComposition.Status.Conditions, condition)
	}

	if err := r.Status().Update(ctx, &customComposition); err != nil {
		log.Error(err, "unable to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func constructPod(customComposition *pocv1alpha1.CustomComposition) *corev1.Pod {
	initContainerVms := []corev1.VolumeMount{
		{
			Name:      "storage",
			MountPath: "/data/",
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customComposition.Name,
			Namespace: customComposition.Namespace,
		},
	}

	pod.Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:         "debug-resources",
				Image:        baseImage,
				Command:      []string{"sh", "-c", "sleep infinity"},
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
	}

	for _, f := range customComposition.Spec.Functions {
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
			Name:  f.Title,
			Image: f.Image,
			Command: []string{"sh", "-c", fmt.Sprintf("kpt fn eval /data --exec /function -- %s",
				f.Args)},
			VolumeMounts: initContainerVms,
		})
	}

	return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomCompositionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pocv1alpha1.CustomComposition{}).
		Complete(r)
}
