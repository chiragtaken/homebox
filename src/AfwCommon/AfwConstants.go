package AfwCommon

var AfwLogger *Logger

const (

	//HTTP REST RESPONSE TYPES
        AfwRestRespSuccess = 0
        AfwRestRespError   = 1
        AfwRestRespWarn    = 2

	//HTTP Server related configs
        HttpListenPort = "35000"
	HttpDir = "/media/pi/Dragon"
)

type AfwUserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AfwToken struct {
	Token string `json:"token"`
}

type AfwRestResponse struct {
        ResponseType int         `json:"ResponseType"`
        Response     interface{} `json:"Response"`
}
