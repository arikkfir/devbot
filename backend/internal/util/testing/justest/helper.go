package justest

import "fmt"

type Helper interface {
	Helper()
}

type HelperProvider interface {
	GetHelper() Helper
}

//go:noinline
func GetHelper(t T) Helper {
	if hp, ok := t.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t))
	}
}
