package originpolymorphichelpers

import (
	"sort"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	api "k8s.io/kubernetes/pkg/apis/core"
	kinternalclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/polymorphichelpers"

	appsapi "github.com/openshift/origin/pkg/apps/apis/apps"
)

func NewAttachablePodForObjectFn(delegate polymorphichelpers.AttachableLogsForObjectFunc) polymorphichelpers.AttachableLogsForObjectFunc {
	return func(restClientGetter genericclioptions.RESTClientGetter, object runtime.Object, timeout time.Duration) (*api.Pod, error) {
		switch t := object.(type) {
		case *appsapi.DeploymentConfig:
			config, err := restClientGetter.ToRESTConfig()
			if err != nil {
				return nil, err
			}
			coreClient, err := kinternalclient.NewForConfig(config)
			if err != nil {
				return nil, err
			}

			selector := labels.SelectorFromSet(t.Spec.Selector)
			f := func(pods []*v1.Pod) sort.Interface {
				return sort.Reverse(controller.ActivePods(pods))
			}
			pod, _, err := polymorphichelpers.GetFirstPod(coreClient.Core(), t.Namespace, selector.String(), 1*time.Minute, f)
			return pod, err

		default:
			return delegate(restClientGetter, object, timeout)
		}
	}
}
