/**
* @Author: TongTongLiu
* @Date: 2021/3/10 11:47 AM
**/

package api

type InterfaceNew interface {
	Invoke()
}

type InterfaceRequest struct {
	MethodName string
	Header     map[string]string
	Body       []byte
	URL        string
}
