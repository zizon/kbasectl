package endpoint

import (
	"net/http"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type Client interface {
	NewRequest(apiPath string, gv schema.GroupVersion) *rest.Request
}

type client struct {
	config rest.Config
	http   *http.Client
	scheme *runtime.Scheme
}

func NewAPIClient(config Config) (Client, error) {
	scheme := runtime.NewScheme()

	config.Rest.APIPath = "/apis"
	config.Rest.GroupVersion = &schema.GroupVersion{}
	config.Rest.NegotiatedSerializer = serializer.NewCodecFactory(scheme)

	rt, err := rest.TransportFor(&config.Rest)

	if err != nil {
		klog.Errorf("fail create client with config: %v reason: %v", config, err)
		return client{}, err
	}
	client := client{
		config: config.Rest,
		http:   &http.Client{Transport: rt},
		scheme: scheme,
	}

	//config.Rest.NegotiatedSerializer = serializer.NewCodecFactory(client.scheme)
	return client, nil
}

func (client client) Register(objType reflect.Type, gv schema.GroupVersion) Client {
	gvk := gv.WithKind(objType.Name())

	metav1.AddToGroupVersion(client.scheme, gv)
	client.scheme.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
	return client
}

func (client client) NewRequest(apiPath string, gv schema.GroupVersion) *rest.Request {
	base, versionurl, err := rest.DefaultServerURL(
		client.config.Host,
		apiPath,
		gv,
		client.config.Insecure,
	)
	if err != nil {
		klog.Errorf("malform request:%v", err)
		return &rest.Request{}
	}

	return rest.NewRequestWithClient(
		base, versionurl, rest.ClientContentConfig{
			AcceptContentTypes: client.config.AcceptContentTypes,
			ContentType:        client.config.ContentType,
			GroupVersion:       gv,
			Negotiator:         runtime.NewClientNegotiator(client.config.NegotiatedSerializer, gv),
		}, client.http,
	)
	//return rest.NewRequest(client.delegate)
}
