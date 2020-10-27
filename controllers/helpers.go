/*
Copyright 2019 Google LLC

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
	"fmt"
	"net"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	vscodev1alpha1 "github.com/naveensrinivasan/api/v1alpha1"
)

func (r *ServerReconciler) desiredDeployment(code vscodev1alpha1.Server) (appsv1.Deployment, error) {
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      code.Name,
			Namespace: code.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"codeserver": code.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"codeserver": code.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "codeserver",
							Image: "codercom/code-server:3.6.1",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080, Name: "http", Protocol: "TCP"},
							},
							Args: []string{
								"--bind-addr", "0.0.0.0:8080", "--auth", "none",
							},
							Resources: *code.Spec.Resources.DeepCopy(),
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&code, &depl, r.Scheme); err != nil {
		return depl, err
	}

	return depl, nil
}

func (r *ServerReconciler) desiredService(code vscodev1alpha1.Server) (corev1.Service, error) {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      code.Name,
			Namespace: code.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromString("http")},
			},
			Selector: map[string]string{"codeserver": code.Name},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	// always set the controller reference so that we know which object owns this.
	if err := ctrl.SetControllerReference(&code, &svc, r.Scheme); err != nil {
		return svc, err
	}

	return svc, nil
}

func urlForService(svc corev1.Service, port int32) string {
	// notice that we unset this if it's not present -- we always want the
	// state to reflect what we observe.
	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return ""
	}

	host := svc.Status.LoadBalancer.Ingress[0].Hostname
	if host == "" {
		host = svc.Status.LoadBalancer.Ingress[0].IP
	}

	return fmt.Sprintf("http://%s", net.JoinHostPort(host, fmt.Sprintf("%v", port)))
}
