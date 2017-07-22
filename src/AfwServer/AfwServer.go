package main

import (
	"AfwCommon"
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

var logger *AfwCommon.Logger
var signingKey, verificationKey []byte

type AfwCustomeUserInfo struct {
	Name string `json:"Name"`
	Role string `json:"Role"`
}

type AfwMyCustomClaims struct {
    UserInfo AfwCustomeUserInfo `json:"AfwCustomeUserInfo"`
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
		token, err := jwt.ParseWithClaims(auth[0], &AfwMyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("All Your_Base_is Covered"), nil
		})	
		if err != nil {
			AfwCommon.AfwSendHttpResponse(resp, http.StatusUnauthorized, AfwCommon.AfwRestRespError, err.Error())
			return
		}
		if claims, ok := token.Claims.(*AfwMyCustomClaims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), MyKey, *claims)
			page(resp, req.WithContext(ctx))
		} else {
			AfwCommon.AfwSendHttpResponse(resp, http.StatusUnauthorized, AfwCommon.AfwRestRespError, err.Error())
			return
		}
	})
}

func LoginHandler(resp http.ResponseWriter, req *http.Request) {

	mySigningKey := []byte("All Your_Base_is Covered")
	var user AfwCommon.AfwUserCredentials

	//decode request into UserCredentials struct
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		AfwCommon.AfwSendHttpResponse(resp, http.StatusInternalServerError, AfwCommon.AfwRestRespError, err.Error())
		return
	}

	logger.Debug(user.Username + " - " + user.Password)

	//validate user credentials
	if user.Username != "ctayal" || user.Password != "Newuser@123" {
		msg := "Error logging in"
		AfwCommon.AfwSendHttpResponse(resp, http.StatusForbidden, AfwCommon.AfwRestRespError, msg)
		return
	}

	expireToken := time.Now().Add(time.Minute * 20).Unix()
	//expireCookie := time.Now().Add(time.Minute * 20)
	claims := AfwMyCustomClaims{
		AfwCustomeUserInfo{Name: user.Username, Role: "Member"},
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
		AfwCommon.AfwSendHttpResponse(resp, http.StatusInternalServerError, AfwCommon.AfwRestRespError, err.Error())
		return
	}

	//create a token instance using the token string
	tok := AfwCommon.AfwToken{tokenString}
	logger.Debug("Token : ", tokenString)

	//cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true}
	//http.SetCookie(res, &cookie)

	AfwCommon.AfwSendHttpResponse(resp, http.StatusOK, AfwCommon.AfwRestRespSuccess, tok)
	return
}

func GetFileHandler(resp http.ResponseWriter, req *http.Request) {
	_, ok := req.Context().Value(MyKey).(AfwMyCustomClaims)
	if !ok {
		msg := "Unauthorized access"
		AfwCommon.AfwSendHttpResponse(resp, http.StatusUnauthorized, AfwCommon.AfwRestRespError, msg)
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
		path = AfwCommon.HttpDir + "/" + path
		logger.Debug("Path: ", path)
		fi, err := os.Stat(path)
		if err != nil {
			AfwCommon.AfwSendHttpResponse(resp, http.StatusInternalServerError, AfwCommon.AfwRestRespError, err.Error())
			return
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			file, _ := ioutil.ReadDir(path)
			for _, f := range file {
				files[f.Name()] = f.Size()
			}
			AfwCommon.AfwSendHttpResponse(resp, http.StatusOK, AfwCommon.AfwRestRespSuccess, files)
			return
		case mode.IsRegular():
			msg := "Cannot open File"
			AfwCommon.AfwSendHttpResponse(resp, http.StatusInternalServerError, AfwCommon.AfwRestRespError, msg)
			return
		}
	} else {
		logger.Debug("Dir: ", AfwCommon.HttpDir)
		file, _ := ioutil.ReadDir(AfwCommon.HttpDir)
		for _, f := range file {
			files[f.Name()] = f.Size()
		}
		logger.Debug("Files: ", files)

		AfwCommon.AfwSendHttpResponse(resp, http.StatusOK, AfwCommon.AfwRestRespSuccess, files)
		return
	}
}

func main() {

	//Make a channel to receive os signals
	sigchan := make(chan os.Signal, 1)

	//Register all the singals that need to be handled by the code
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	AfwCommon.AfwLogger = new(AfwCommon.Logger)
	AfwCommon.AfwLogger.Init("/var/log/afwserver.log")
	AfwCommon.AfwLogger.SetLoglevel("DEBUG")
	logger = AfwCommon.AfwLogger

	logger.Debug("Starting Application REST Router\n")
	http.HandleFunc("/afw/login", LoginHandler)
	http.HandleFunc("/afw/list", validate(GetFileHandler))


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
