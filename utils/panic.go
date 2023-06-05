package utils

// shortcut for
//
//	if err != nil {
//	    panic(err)
//	}
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
