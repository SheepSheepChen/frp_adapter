
package noticetrigger

import (
	"math"
	"time"
)

/*告警触发器*/
type NoticeTrigger struct {
	//发送上线消息的时间
	ConnectTime time.Time
	//发送短线消息的时间
	DisconnectTime time.Time

}

//更新发送上线消息的时间
func (nt *NoticeTrigger)  UpdateConnectTime(){
	nt.ConnectTime=time.Now()
}
//发送下线消息的时间
func (nt *NoticeTrigger)  UpdateDieconnectTime(){
	nt.DisconnectTime=time.Now()
}

func  (nt *NoticeTrigger) IsNotice() bool{
	subTime :=nt.ConnectTime.Sub(nt.DisconnectTime)
	if math.Abs(subTime.Seconds())>30.0{
		return true
	}
	return false
}




