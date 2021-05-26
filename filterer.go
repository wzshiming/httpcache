package httpcache

import (
	"net/http"
	"strings"
)

func AndJointFilterer(filterers ...Filterer) Filterer {
	switch len(filterers) {
	case 0:
		return nil
	case 1:
		return filterers[0]
	}
	return FiltererFunc(func(req *http.Request) bool {
		for _, filterer := range filterers {
			if !filterer.Filter(req) {
				return false
			}
		}
		return true
	})
}

func OrJointFilterer(filterers ...Filterer) Filterer {
	switch len(filterers) {
	case 0:
		return nil
	case 1:
		return filterers[0]
	}
	return FiltererFunc(func(req *http.Request) bool {
		for _, filterer := range filterers {
			if filterer.Filter(req) {
				return true
			}
		}
		return false
	})
}

func MethodFilterer(m string) Filterer {
	return FiltererFunc(func(req *http.Request) bool {
		return req.Method == m
	})
}

func PrefixFilterer(prefix string) Filterer {
	return FiltererFunc(func(req *http.Request) bool {
		return strings.HasPrefix(req.URL.Path, prefix)
	})
}
