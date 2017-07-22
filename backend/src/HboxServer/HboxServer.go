package main

import (
	"HboxCommon"
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"context"
	"encoding/json"
	"encoding/pem"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Key int
const MyKey Key = 0

var logger *HboxCommon.Logger
var signingKey, verificationKey []byte

type HboxCustomeUserInfo struct {
	Name string `json:"Name"`
	Role string `json:"Role"`
}

type HboxMyCustomClaims struct {
    UserInfo HboxCustomeUserInfo `json:"HboxCustomeUserInfo"`
    jwt.StandardClaims
}

func initKeys() {

	key, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		logger.Fatal("Error generating private key")
                os.Exit(1)
	}
	publicKey := key.PublicKey 

	var privPEMBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key), 
	}
	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, privPEMBlock)
	
	signingKey = privKeyPEMBuffer.Bytes()

	/* create verificationKey from pubKey. Also in PEM-format */
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey) 
	if err != nil {
		logger.Fatal("Error marshalling public key")
                os.Exit(1)
	}

	var pubPEMBlock = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	}
	pubKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(pubKeyPEMBuffer, pubPEMBlock)

	verificationKey = pubKeyPEMBuffer.Bytes()
}

func validate(page http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		auth := req.Header["Authorization"]
		token, err := jwt.ParseWithClaims(auth[0], &HboxMyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("All Your_Base_is Covered"), nil
		})	
		if err != nil {
			HboxCommon.HboxSendHttpResponse(resp, http.StatusUnauthorized, HboxCommon.HboxRestRespError, err.Error())
			return
		}
		if claims, ok := token.Claims.(*HboxMyCustomClaims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), MyKey, *claims)
			page(resp, req.WithContext(ctx))
		} else {
			HboxCommon.HboxSendHttpResponse(resp, http.StatusUnauthorized, HboxCommon.HboxRestRespError, err.Error())
			return
		}
	})
}

func LoginHandler(resp http.ResponseWriter, req *http.Request) {

	mySigningKey := []byte("All Your_Base_is Covered")
	var user HboxCommon.HboxUserCredentials

	//decode request into UserCredentials struct
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		HboxCommon.HboxSendHttpResponse(resp, http.StatusInternalServerError, HboxCommon.HboxRestRespError, err.Error())
		return
	}

	logger.Debug(user.Username + " - " + user.Password)

	//validate user credentials
	if user.Username != "ctayal" || user.Password != "Newuser@123" {
		msg := "Error logging in"
		HboxCommon.HboxSendHttpResponse(resp, http.StatusForbidden, HboxCommon.HboxRestRespError, msg)
		return
	}

	expireToken := time.Now().Add(time.Minute * 20).Unix()
	//expireCookie := time.Now().Add(time.Minute * 20)
	claims := HboxMyCustomClaims{
		HboxCustomeUserInfo{Name: user.Username, Role: "Member"},
		jwt.StandardClaims{
			Issuer: "admin",
			ExpiresAt: expireToken,
			IssuedAt: time.Now().Unix(),
			Id: "1001",
		},
	}

	//create a rsa 256 signer
	signer := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := signer.SignedString(mySigningKey)
	if err != nil {
		logger.Error("Error while signing the token")
		HboxCommon.HboxSendHttpResponse(resp, http.StatusInternalServerError, HboxCommon.HboxRestRespError, err.Error())
		return
	}

	//create a token instance using the token string
	tok := HboxCommon.HboxToken{tokenString}
	logger.Debug("Token : ", tokenString)

	//cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true}
	//http.SetCookie(res, &cookie)

	HboxCommon.HboxSendHttpResponse(resp, http.StatusOK, HboxCommon.HboxRestRespSuccess, tok)
	return
}

func GetFileHandler(resp http.ResponseWriter, req *http.Request) {
	_, ok := req.Context().Value(MyKey).(HboxMyCustomClaims)
	if !ok {
		msg := "Unauthorized access"
		HboxCommon.HboxSendHttpResponse(resp, http.StatusUnauthorized, HboxCommon.HboxRestRespError, msg)
		return
	}

	queryMap := req.URL.Query()
	var path string
	var files = make(map[string]int64)

	if len(queryMap) > 0 {
		if val, exists := queryMap["path"]; exists {
                        path = strings.Trim(val[0], " ")
                }
	}

	if path != "" {
		path = HboxCommon.HttpDir + "/" + path
		logger.Debug("Path: ", path)
		fi, err := os.Stat(path)
		if err != nil {
			HboxCommon.HboxSendHttpResponse(resp, http.StatusInternalServerError, HboxCommon.HboxRestRespError, err.Error())
			return
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			file, _ := ioutil.ReadDir(path)
			for _, f := range file {
				files[f.Name()] = f.Size()
			}
			HboxCommon.HboxSendHttpResponse(resp, http.StatusOK, HboxCommon.HboxRestRespSuccess, files)
			return
		case mode.IsRegular():
			msg := "Cannot open File"
			HboxCommon.HboxSendHttpResponse(resp, http.StatusInternalServerError, HboxCommon.HboxRestRespError, msg)
			return
		}
	} else {
		logger.Debug("Dir: ", HboxCommon.HttpDir)
		file, _ := ioutil.ReadDir(HboxCommon.HttpDir)
		for _, f := range file {
			files[f.Name()] = f.Size()
		}
		logger.Debug("Files: ", files)

		HboxCommon.HboxSendHttpResponse(resp, http.StatusOK, HboxCommon.HboxRestRespSuccess, files)
		return
	}
}

func main() {

	//Make a channel to receive os signals
	sigchan := make(chan os.Signal, 1)

	//Register all the singals that need to be handled by the code
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	HboxCommon.HboxLogger = new(HboxCommon.Logger)
	HboxCommon.HboxLogger.Init("/var/log/hboxserver.log")
	HboxCommon.HboxLogger.SetLoglevel("DEBUG")
	logger = HboxCommon.HboxLogger

	logger.Debug("Starting Application REST Router\n")
	http.HandleFunc("/hbox/login", LoginHandler)
	http.HandleFunc("/hbox/list", validate(GetFileHandler))


	http.ListenAndServe(":35000", nil)

	//Block until a signal is received
	for {
		switch sigrcvd := <-sigchan; sigrcvd {
		case syscall.SIGINT, syscall.SIGTERM:
			//TODO What cleanup is needed ? Call destructors of each modules?
			logger.Debug("Exiting due to singnal %d", sigrcvd)
			os.Exit(0)
		}
	}
}
