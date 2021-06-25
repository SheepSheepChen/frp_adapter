package notification

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ttlv/frp_adapter/frp_adapter_init"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/kubernetes"
)

func EventNotice(c *gin.Context) {
	//新建clientset客户端
	clientset, err := frp_adapter_init.NewClientset()
	if err != nil {
		logrus.Error(err)

	}
	//获取deploy uid
	deploy ,err :=clientset.AppsV1().Deployments("default").Get("frp-adapter",metav1.GetOptions{})
	if err != nil {
		logrus.Error(err)

	}
	//新建 event对象
	event :=&apiv1.Event{}
	if c.Request.FormValue("status") == "offline" {
		event = &apiv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name: "frp-adapter-node-"+c.Request.FormValue("unique_id")+"."+rand.String(8),
			},
			InvolvedObject: apiv1.ObjectReference{
				Kind:      "deployment",
				Name:      "frp-adapter",
				Namespace: "default",
				UID: deploy.ObjectMeta.UID,
			},
			FirstTimestamp: metav1.Time{time.Now()},
			Message:        "node-" +c.Request.FormValue("unique_id")+" disconnect",
			Reason:         "FRP disconnect",
			Type:           "Warning",
		}
	}else if c.Request.FormValue("status") == "online"{
		event = &apiv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name: "frp-adapter-node-"+c.Request.FormValue("unique_id")+"."+rand.String(10),
			},
			InvolvedObject: apiv1.ObjectReference{
				Kind:      "deployment",
				Name:      "frp-adapter",
				Namespace: "default",
				UID: deploy.ObjectMeta.UID,
			},
			FirstTimestamp: metav1.Time{time.Now()},
			Message:        "node-" +c.Request.FormValue("unique_id")+" connect",
			Reason:         "FRP connect",
			Type:           "Normal",
		}
	}

	//创建event
	result, err := clientset.CoreV1().Events("default").Create(event)
	if err != nil {
		logrus.Error(err)

	}
	fmt.Printf("event:%v\n", result.Message)

}
