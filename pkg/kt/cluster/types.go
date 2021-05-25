package cluster

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Create kubernetes instance
func Create(kubeConfig string) (kubernetes KubernetesInterface, err error) {
	clientSet, err := getKubernetesClient(kubeConfig)
	if err != nil {
		return
	}
	return &Kubernetes{
		Clientset: clientSet,
	}, nil
}

// CreateFromClientSet kubernetes instance
func CreateFromClientSet(clientSet kubernetes.Interface) (kubernetes KubernetesInterface, err error) {
	return &Kubernetes{
		Clientset: clientSet,
	}, nil
}

// KubernetesInterface kubernetes interface
type KubernetesInterface interface {
	RemoveDeployment(name, namespace string) (err error)
	RemoveConfigMap(name, namespace string) (err error)
	RemoveService(name, namespace string) (err error)
	Deployment(name, namespace string) (deployment *appV1.Deployment, err error)
	Scale(deployment *appV1.Deployment, replicas *int32) (err error)
	ScaleTo(deployment, namespace string, replicas *int32) (err error)
	//获取该命名空间下所有服务的 服务名称及ip地址映射
	ServiceHosts(namespace string) (hosts map[string]string)
	ClusterCrids(namespace string, podCIDR string) (cidrs []string, err error)
	//获取或创建pod
	GetOrCreateShadow(name, namespace, image string, labels, envs map[string]string, debug, reuseShadow bool) (podIP, podName, sshcm string, credential *util.SSHCredential, err error)
	//创建服务
	CreateService(name, namespace string, port int, labels map[string]string) (*coreV1.Service, error)
	GetDeployment(name string, namespace string) (*appV1.Deployment, error)
	UpdateDeployment(namespace string, deployment *appV1.Deployment) (*appV1.Deployment, error)
	DecreaseRef(namespace string, deployment string) (cleanup bool, err error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	KubeConfig string
	Clientset  kubernetes.Interface
}
