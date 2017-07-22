/*
 *    Copyright (c) 2017 by Cisco Systems, Inc.
 *    All rights reserved.
 */
package AfwCommon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type RestRouteHandler func(http.ResponseWriter, *http.Request)

type restRoute struct {
	match    string
	methods  []string
	handler  RestRouteHandler
	children map[string]restRoute
}

var registeredRoutes map[string]restRoute

type afwAppRestDispatcher struct {
	handler http.Handler
}

func afwRestIsValidMethodForRouter(rt restRoute, method string) bool {

	AfwLogger.Debugf("Checking Method %s %v\n", method, rt)
	for _, v := range rt.methods {
		if v == method {
			return true
		}
	}
	return false
}

func (x afwAppRestDispatcher) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {

	AfwLogger.Infof("Rest Call Received for URL %s Method %s\n", req.URL.Path, req.Method)

	routes := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	router := afwRestMatchRoute(routes, registeredRoutes)
	if router == nil {
		AfwLogger.Error("Cannot Find a Valid Router for the Rest Request")
		rsp.WriteHeader(http.StatusNotFound)
		return
	}

	if !afwRestIsValidMethodForRouter(*router, req.Method) {
		AfwLogger.Error("Cannot Find a Valid Method in Router for the Rest Request")
		rsp.WriteHeader(http.StatusNotFound)
		return
	}

	if router.handler != nil {
		router.handler(rsp, req)
	}
}

func afwRestMatchRoute(routes []string, rt map[string]restRoute) *restRoute {

	matchrt, matched := rt[routes[0]]
	if !matched {
		for _, matchrt = range rt {
			if !strings.Contains(matchrt.match, ".*[]{}-$|,^") {
				AfwLogger.Debugf("This one is a regular expression %s\n", matchrt.match)
				matched, _ = regexp.MatchString(matchrt.match, routes[0])
				if matched {
					break
				}
			}
		}
	}

	if !matched {
		return nil
	}

	AfwLogger.Debugf("This Route is %s, %d\n", routes[0], len(routes))

	if len(routes) == 1 {
		return &matchrt
	}

	x := afwRestMatchRoute(routes[1:len(routes)], matchrt.children)
	//	if x != nil {
	//		fmt.Printf("Returning %v", x)
	//	}
	return x
}

func AfwRegisterRestRoute(route string, methods []string, handler RestRouteHandler) error {

	var k, l int
	var v string
	var rt map[string]restRoute
	var leafrt restRoute

	AfwLogger.Debugf("Registering new Route %s\n", route)
	routes := strings.Split(route, "/")

	/* Find the bottom where you need to start inserting the new route from */
	rt = registeredRoutes
	for k, v = range routes {
		_, ok := rt[v]
		if !ok {
			AfwLogger.Debugf("Found Match at %d Route %s\n", k, v)
			break
		}
		leafrt = rt[v]
		rt = leafrt.children
	}

	if k == len(routes) {
		return errors.New("Duplicate Route Registration")
	}

	for l = k; l < len(routes); l++ {
		r := restRoute{routes[l], nil, nil, make(map[string]restRoute)}
		if l == (len(routes) - 1) {
			r.handler = handler
			r.methods = methods
		}
		leafrt.children[routes[l]] = r
		leafrt = r
	}

	AfwLogger.Debugf("Added Leaf Node %s %v\n", leafrt.match, leafrt.methods)
	return nil
}

func AfwStartRestRouter() {
	registeredRoutes = make(map[string]restRoute)
	registeredRoutes["afw"] = restRoute{"afw", []string{"GET"}, nil, make(map[string]restRoute)}
	go func() {
		http.ListenAndServe(":"+HttpListenPort, afwAppRestDispatcher{})
		AfwLogger.Fatal("Unable to bind To HTTP Port")
		os.Exit(1)
	}()
}

func SendHttpResponse(resp http.ResponseWriter, status int, jsonObj interface{}) {
	resp.WriteHeader(status)
	resp.Header().Set("Content-Type", "application/json")
	jsonB, _ := json.Marshal(jsonObj)
	fmt.Fprintf(resp, string(jsonB))
}

func AfwSendHttpResponse(resp http.ResponseWriter, status int, responseType int, response interface{}) {
	httpResp := AfwRestResponse{}
	httpResp.ResponseType = responseType
	httpResp.Response = response
	SendHttpResponse(resp, status, httpResp)
}
