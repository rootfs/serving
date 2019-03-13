package main

import (
	"github.com/knative/serving/lib"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/go-openapi/spec"
	"github.com/golang/glog"

	serving_install "github.com/knative/serving/pkg/apis/serving/install"
	serving_v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
)

func main() {

	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	serving_install.Install(Scheme)

	mapper := meta.NewDefaultRESTMapper(nil)
	mapper.AddSpecific(serving_v1alpha1.SchemeGroupVersion.WithKind("Revision"),
		serving_v1alpha1.SchemeGroupVersion.WithResource("revisions"),
		serving_v1alpha1.SchemeGroupVersion.WithResource("revision"), meta.RESTScopeRoot)

	spec, err := lib.RenderOpenAPISpec(lib.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "Knative Serving",
			Version: "v0.1",
			Contact: &spec.ContactInfo{
				Name: "Knative Serving",
				URL:  "https://knative.io/",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			serving_v1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []schema.GroupVersionResource{
			serving_v1alpha1.SchemeGroupVersion.WithResource("Revision"),
		},
		Mapper: mapper,
	})
	if err != nil {
		glog.Info(err)
	}

	filename := "swagger.json"
	err = ioutil.WriteFile(filename, []byte(spec), 0644)
	if err != nil {
		glog.Fatal(err)
	}
}
