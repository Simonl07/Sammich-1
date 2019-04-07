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
		"Show",
		"GET",
		"/view/chain",
		Show,
	},
	Route{
		"RegisterBusiness",
		"POST",
		"/register",
		RegisterBusiness,
	},
	Route{
		"Accept",
		"POST",
		"/accept/{company}/{uid}",
		Accept,
	},
	Route{
		"ViewCache",
		"GET",
		"/cache",
		ViewCache,
	},
	Route{
		"FetchMerits",
		"GET",
		"/view/merits",
		FetchMerits,
	},
	Route{
		"FetchAcceptances",
		"GET",
		"/view/acceptances",
		FetchAcceptances,
	},
	Route{
		"Download",
		"GET",
		"/download",
		Download,
	},
	Route{
		"ShowKeys",
		"GET",
		"/keys",
		ShowKeys,
	},
}
