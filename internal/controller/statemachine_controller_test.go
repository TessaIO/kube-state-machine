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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubetessaiov1 "github.com/TessaIO/kube-state-machine/api/v1"
)

var _ = Describe("StateMachine Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "fsm"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		timeout := time.Second * 10
		interval := time.Millisecond * 250

		AfterEach(func() {
			resource := &kubetessaiov1.StateMachine{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance StateMachine")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the task state machine", func() {
			By("By creating a new state machine with a task state")
			ctx := context.Background()

			stateMachine := &kubetessaiov1.StateMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: kubetessaiov1.StateMachineSpec{
					States: []kubetessaiov1.State{
						{
							Name: "state1",
							Type: kubetessaiov1.StateTypeTask,
							Task: &corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test-container",
										Image: "test-image",
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, stateMachine)).Should(Succeed())

			stateMachineLookupKey := types.NamespacedName{Name: resourceName, Namespace: "default"}
			createdStateMachine := &kubetessaiov1.StateMachine{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, stateMachineLookupKey, createdStateMachine)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(len(createdStateMachine.Spec.States)).Should(Equal(1))
			Expect(createdStateMachine.Spec.States[0].Type).Should(Equal(kubetessaiov1.StateTypeTask))
		})
	})
})
