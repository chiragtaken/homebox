package HboxCommon

var HboxLogger *Logger

const (

	//HTTP REST RESPONSE TYPES
        HboxRestRespSuccess = 0
        HboxRestRespError   = 1
        HboxRestRespWarn    = 2

	//HTTP Server related configs
        HttpListenPort = "35000"
	HttpDir = "/media/pi/Dragon"
)

type HboxUserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HboxToken struct {
	Token string `json:"token"`
}

type HboxRestResponse struct {
        ResponseType int         `json:"ResponseType"`
        Response     interface{} `json:"Response"`
}
