package v1beta1

import corev1 "k8s.io/api/core/v1"

func (e *EnvironmentSpec) ContainerPorts() []corev1.ContainerPort {
	ports := []corev1.ContainerPort{}
	for _, p := range e.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          p.Name,
			ContainerPort: p.Port,
			Protocol:      p.Protocol,
		})
	}

	return ports
}
