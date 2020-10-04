package defaults

import (
	"bytes"

	"github.com/zizon/kbasectl/pkg/panichain"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	KbaseCTLNamespace  = "kbase"
	KbaseCephTokenName = "ceph-kbase-token"
)

func KubernatesObjectsToYaml(objects []runtime.Object) []byte {
	buf := bytes.Buffer{}
	for _, obj := range objects {
		annotateWithVersion(obj)
		buf.WriteString("---\n")
		if out, err := yaml.Marshal(obj); err != nil {
			panichain.Propogate(err)
		} else if _, err := buf.Write(out); err != nil {
			panichain.Propogate(err)
		}
		buf.WriteString("...\n")
	}

	return buf.Bytes()
}
