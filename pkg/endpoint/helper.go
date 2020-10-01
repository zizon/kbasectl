package endpoint

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/zizon/kbasectl/pkg/panichain"
	"k8s.io/klog"
)

type NamesapcedObject struct {
	Group     string
	Version   string
	Kind      string
	Namespace string
	API       string
	Name      string
	Object    interface{}
}

func MaybeCreate(client Client, obj NamesapcedObject) {
	basepath := path.Join("/", obj.API, obj.Group, obj.Version)
	if len(obj.Namespace) > 0 {
		basepath = path.Join(basepath, "namespaces", obj.Namespace, obj.Kind)
	} else {
		basepath = path.Join(basepath, obj.Kind)
	}

	if klog.V(5) {
		klog.Infof("base url:%s", basepath)
	}

	response := client.Do(http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: path.Join(basepath, obj.Name),
		},
	})

	if response.StatusCode == http.StatusNotFound {
		response = client.Do(http.Request{
			Method: "POST",
			URL: &url.URL{
				Path: basepath,
			},
			Body: NewRequestBody(obj.Object),
		})

		switch response.StatusCode {
		case http.StatusCreated, http.StatusConflict:
			//pass
		default:
			panichain.Propogate(fmt.Errorf("fail ensure exist: %s/%s %v", basepath, obj.Name, response))
			return
		}
	}
}
