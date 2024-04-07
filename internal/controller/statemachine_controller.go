/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubetessaiov1 "github.com/TessaIO/kube-state-machine/api/v1"
)

// StateMachineReconciler reconciles a StateMachine object
type StateMachineReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	stateMachineManager *StateMachineManager
}

//+kubebuilder:rbac:groups=kube.tessa.io.kube.state.machine,resources=statemachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube.tessa.io.kube.state.machine,resources=statemachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube.tessa.io.kube.state.machine,resources=statemachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StateMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *StateMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	stateMachine := &kubetessaiov1.StateMachine{}
	if err := r.Get(ctx, req.NamespacedName, stateMachine); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// FIXME: We should change the current state from init state to another state based on the state machine status
	r.stateMachineManager = NewStateMachineManager(InitState, stateMachine.Spec.States)
	if err := r.ReconcileStates(ctx, stateMachine); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *StateMachineReconciler) ReconcileStates(ctx context.Context, stateMachine *kubetessaiov1.StateMachine) error {
	log := log.FromContext(ctx)
	log.Info("Started reconciling states")
loop:
	for s := r.stateMachineManager.currentState; s != nil; s = r.stateMachineManager.Next() {
		if *s == InitState {
			log.Info("Skipping init state")
			continue
		}

		state := getStateByName(*s, stateMachine)
		if state == nil {
			return fmt.Errorf("state %s not found", *s)
		}

		log.Info(fmt.Sprintf("Reconciling state %s, type: %v", state.Name, state.Type))
		switch state.Type {
		case kubetessaiov1.StateTypeTask:
			// Here we should deploy the workload
			job := createJob(*stateMachine, state.Task)
			err := r.Client.Create(ctx, &job)
			if err != nil {
				log.Error(err, "error while creating job", "state", state.Name)
				return err
			}
		case kubetessaiov1.StateTypePass:
			continue
		// TODO: add update status of CRD in the future
		case kubetessaiov1.StateTypeFail:
			break loop
		case kubetessaiov1.StateTypeWait:
			if state.WaitFor == nil {
				log.Info("waitFor attribute is not specified, skipping...")
				continue
			}
			duration, err := time.ParseDuration(*state.WaitFor)
			if err != nil {
				log.Error(err, "error parsing duration", "state", state.Name)
				return err
			}
			time.Sleep(duration)
		}
	}
	return nil
}

func createJob(stateMachine kubetessaiov1.StateMachine, podSpec *corev1.PodSpec) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stateMachine.Name,
			Namespace: stateMachine.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: *podSpec,
			},
		},
	}
}

func getStateByName(state string, stateMachine *kubetessaiov1.StateMachine) *kubetessaiov1.State {
	for _, s := range stateMachine.Spec.States {
		if s.Name == state {
			return &s
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StateMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubetessaiov1.StateMachine{}).
		Complete(r)
}
