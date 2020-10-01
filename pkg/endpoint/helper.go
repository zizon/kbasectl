package endpoint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"

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

func MaybeUpdate(client Client, update func(NamesapcedObject) NamesapcedObject) {
	obj := update(NamesapcedObject{Object: nil})

	basepath := path.Join("/", obj.API, obj.Group, obj.Version)
	if len(obj.Namespace) > 0 {
		basepath = path.Join(basepath, "namespaces", obj.Namespace, obj.Kind)
	} else {
		basepath = path.Join(basepath, obj.Kind)
	}

	if klog.V(5) {
		klog.Infof("base url:%s", basepath)
	}

	// 1. get existing one, if presented
	response := client.Do(http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: path.Join(basepath, obj.Name),
		},
	})
	if response.StatusCode == http.StatusNotFound {
		if klog.V(5) {
			klog.Infof("try createing object:%s", obj.Object)
		}
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
			body, _ := ioutil.ReadAll(response.Body)
			panichain.Propogate(fmt.Errorf("fail to craete: %s/%s %v\n%s",
				basepath,
				obj.Name,
				response.Status,
				string(body),
			))
			return
		}
	}

	// 2. decode response
	fetched := reflect.New(reflect.TypeOf(obj.Object)).Interface()
	panichain.Propogate(json.NewDecoder(response.Body).Decode(fetched))

	// 3. call user udpate/merve
	relicated := *&obj
	relicated.Object = fetched
	updated := update(relicated)

	if klog.V(5) {
		klog.Infof("object before update:%v to:%v", fetched, updated.Object)
	}

	// 4. update
	if updated != relicated {
		response = client.Do(http.Request{
			Method: "PUT",
			URL: &url.URL{
				Path: basepath,
			},
			Body: NewRequestBody(updated.Object),
		})
	}
}
