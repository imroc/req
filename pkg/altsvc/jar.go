package altsvc

type Jar interface {
	SetAltSvc(addr string, as *AltSvc)
	GetAltSvc(addr string) *AltSvc
}
