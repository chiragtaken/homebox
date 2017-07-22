/*
 *    Copyright (c) 2017 by Cisco Systems, Inc.
 *    All rights reserved.
 */
package HboxCommon

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

type hboxAppRestDispatcher struct {
	handler http.Handler
}

func hboxRestIsValidMethodForRouter(rt restRoute, method string) bool {

	HboxLogger.Debugf("Checking Method %s %v\n", method, rt)
	for _, v := range rt.methods {
		if v == method {
			return true
		}
	}
	return false
}

func (x hboxAppRestDispatcher) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {

	HboxLogger.Infof("Rest Call Received for URL %s Method %s\n", req.URL.Path, req.Method)

	routes := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	router := hboxRestMatchRoute(routes, registeredRoutes)
	if router == nil {
		HboxLogger.Error("Cannot Find a Valid Router for the Rest Request")
		rsp.WriteHeader(http.StatusNotFound)
		return
	}

	if !hboxRestIsValidMethodForRouter(*router, req.Method) {
		HboxLogger.Error("Cannot Find a Valid Method in Router for the Rest Request")
		rsp.WriteHeader(http.StatusNotFound)
		return
	}

	if router.handler != nil {
		router.handler(rsp, req)
	}
}

func hboxRestMatchRoute(routes []string, rt map[string]restRoute) *restRoute {

	matchrt, matched := rt[routes[0]]
	if !matched {
		for _, matchrt = range rt {
			if !strings.Contains(matchrt.match, ".*[]{}-$|,^") {
				HboxLogger.Debugf("This one is a regular expression %s\n", matchrt.match)
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

	HboxLogger.Debugf("This Route is %s, %d\n", routes[0], len(routes))

	if len(routes) == 1 {
		return &matchrt
	}

	x := hboxRestMatchRoute(routes[1:len(routes)], matchrt.children)
	//	if x != nil {
	//		fmt.Printf("Returning %v", x)
	//	}
	return x
}

func HboxRegisterRestRoute(route string, methods []string, handler RestRouteHandler) error {

	var k, l int
	var v string
	var rt map[string]restRoute
	var leafrt restRoute

	HboxLogger.Debugf("Registering new Route %s\n", route)
	routes := strings.Split(route, "/")

	/* Find the bottom where you need to start inserting the new route from */
	rt = registeredRoutes
	for k, v = range routes {
		_, ok := rt[v]
		if !ok {
			HboxLogger.Debugf("Found Match at %d Route %s\n", k, v)
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

	HboxLogger.Debugf("Added Leaf Node %s %v\n", leafrt.match, leafrt.methods)
	return nil
}

func HboxStartRestRouter() {
	registeredRoutes = make(map[string]restRoute)
	registeredRoutes["hbox"] = restRoute{"hbox", []string{"GET"}, nil, make(map[string]restRoute)}
	go func() {
		http.ListenAndServe(":"+HttpListenPort, hboxAppRestDispatcher{})
		HboxLogger.Fatal("Unable to bind To HTTP Port")
		os.Exit(1)
	}()
}

func SendHttpResponse(resp http.ResponseWriter, status int, jsonObj interface{}) {
	resp.WriteHeader(status)
	resp.Header().Set("Content-Type", "application/json")
	jsonB, _ := json.Marshal(jsonObj)
	fmt.Fprintf(resp, string(jsonB))
}

func HboxSendHttpResponse(resp http.ResponseWriter, status int, responseType int, response interface{}) {
	httpResp := HboxRestResponse{}
	httpResp.ResponseType = responseType
	httpResp.Response = response
	SendHttpResponse(resp, status, httpResp)
}
