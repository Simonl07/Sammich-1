package p3

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Apply",
		"POST",
		"/apply",
		Apply,
	},
	Route{
		"FetchApplications",
		"GET",
		"/fetch",
		FetchApplications,
	},
	Route{
		"RegisterBusiness",
		"POST",
		"/register/{company}",
		RegisterBusiness,
	},
	Route{
		"Accept",
		"POST",
		"/accept/{company}/{uid}",
		Accept,
	},
	Route{
		"Show",
		"GET",
		"/show",
		Show,
	},
	Route{
		"Download",
		"GET",
		"/download",
		Download,
	},
}
