package server

func assert(cond bool, msg string) {
	if !cond {
		panic("Assertion failed: " + msg)
	}
}
