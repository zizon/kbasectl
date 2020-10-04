package defaults

import (
	"github.com/zizon/kbasectl/pkg/panichain"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	metav1.AddMetaToScheme(scheme)
	v1.AddToScheme(scheme)
	appv1.AddToScheme(scheme)
}

func annotateWithVersion(obj runtime.Object) {
	if gvk, _, err := scheme.ObjectKinds(obj); err != nil {
		panichain.Propogate(err)
	} else {
		obj.GetObjectKind().SetGroupVersionKind(gvk[0])
	}

	return
}
