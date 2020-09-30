package endpoint

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type AutoObject struct {
	gvk      schema.GroupVersionKind
	instance interface{}
}

func (obj AutoObject) GetObjectKind() schema.ObjectKind {
	return &obj
}

func (obj AutoObject) DeepCopyObject() runtime.Object {
	copy := *&obj
	return copy
}

func (obj AutoObject) GroupVersionKind() schema.GroupVersionKind {
	return obj.gvk
}

func (obj *AutoObject) SetGroupVersionKind(kind schema.GroupVersionKind) {
	obj.gvk = kind
}
