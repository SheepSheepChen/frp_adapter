package action

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"net/http"

	"fmt"
	"github.com/gorilla/sessions"
	"github.com/ttlv/frp_adapter/app/helpers"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Handlers struct {
	SessionStore  *sessions.CookieStore
	DynamicClient dynamic.Interface
	NameSpace     string
	Res           schema.GroupVersionResource
}

func NewHandlers(sessionStore *sessions.CookieStore, dynamicClient dynamic.Interface, nameSpace string, res schema.GroupVersionResource) Handlers {
	return Handlers{SessionStore: sessionStore, DynamicClient: dynamicClient, NameSpace: nameSpace, Res: res}
}

// 当有新的frpc注册时立即创建新的nodemaintenances对象,优先判断当前nodemaintenances对象是否存在，如果存在则不创建
func (handler *Handlers) FrpCreate(w http.ResponseWriter, r *http.Request) {
	result, getErr := handler.DynamicClient.Resource(handler.Res).Namespace(handler.NameSpace).Get(fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id")), metav1.GetOptions{})
	if getErr != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("failed to get latest version of NodeMaintenance: %v", getErr))
		return
	}
	if result != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("%v is already exist", fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id"))))
		return
	}
	nodeMaintenance := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ke.harmonycloud.io/v1",
			"kind":       "NodeMaintenance",
			"metadata": map[string]interface{}{
				"name":       fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id")),
				"labels":     map[string]interface{}{},
				"annotation": map[string]interface{}{},
			},
			"spec": map[string]interface{}{
				"nodeName": fmt.Sprintf("node-%v", r.FormValue("unique_id")),
				"proxy": map[string]interface{}{
					"type": "FRP",
				},
				"services": map[string]interface{}{
					"name":               fmt.Sprintf("ssh-%v", r.FormValue("unique_id")),
					"type":               "ssh",
					"proxyPort":          r.FormValue("port"),
					"frpServerIpAddress": r.FormValue("frp_server_ip_address"),
					"uniqueID":           r.FormValue("unique_id"),
				},
			},
			"status": map[string]interface{}{
				"services": map[string]interface{}{
					"name":       fmt.Sprintf("ssh-%v", r.FormValue("unique_id")),
					"status":     r.FormValue("status"),
					"lastUpdate": time.Now().UTC().Format("2006-01-02 15:04:05"),
				},
				"conditions": map[string]interface{}{
					"name":      "Maintenable",
					"status":    true,
					"timeStamp": fmt.Sprintf("%v", time.Now().Unix()),
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating NodeMaintenance...")
	result, err := handler.DynamicClient.Resource(handler.Res).Namespace(handler.NameSpace).Create(nodeMaintenance, metav1.CreateOptions{})
	if err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
	helpers.RenderSuccessJSON(w, 200, "Frp client info is created into k8s successfully")
}

// 当frpc的状态更新时需要立即更新nodemaintenances资源
func (handler *Handlers) FrpUpdate(w http.ResponseWriter, r *http.Request) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := handler.DynamicClient.Resource(handler.Res).Namespace(handler.NameSpace).Get(r.FormValue("node_maintenance_name"), metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get latest version of NodeMaintenance: %v", getErr)
		}

		// update Status
		services, found, err := unstructured.NestedMap(result.Object, "status", "services")
		if err != nil || !found || services == nil {
			return fmt.Errorf("nodemaintenance services not found or error in spec: %v", err)
		}
		if err := unstructured.SetNestedField(services, r.FormValue("status"), "status"); err != nil {
			return fmt.Errorf("SetNestedField error: %v", err)
		}

		_, updateErr := handler.DynamicClient.Resource(handler.Res).Namespace(handler.NameSpace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("update failed: %v", retryErr))
		return
	}
}
