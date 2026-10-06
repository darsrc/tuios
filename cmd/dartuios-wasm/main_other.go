//go:build !(js && wasm)

// Command dartuios-wasm is the browser build of dartuios. It only does something when
// built for js/wasm; see build.sh.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "dartuios-wasm runs in a browser. Build it with cmd/dartuios-wasm/build.sh.")
	os.Exit(2)
}
