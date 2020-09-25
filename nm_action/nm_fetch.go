package nm_action

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"strings"
)

func NMFetch(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) (nms []string, err error) {
	lists, err := dynamicClient.Resource(gvr).List(metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf(fmt.Sprintf("NM fetch failed, err is: %v"))
		return
	}
	for _, list := range lists.Items {
		nmNmae, found, err := unstructured.NestedString(list.Object, "metadata", "name")
		if nmNmae == "" || !found || err != nil {
			err = fmt.Errorf(fmt.Sprintf("NM fetch failed, err is: %v"))
			return
		}
		nms = append(nms, strings.Split(nmNmae, "-")[1])
	}
	return
}