//go:build ignore

package main

import "github.com/coalaura/plain/minimal"

func main() {
	log := minimal.New()

	log.Infoln("building release")
	log.Subln("generated configuration")
	log.Subln("compiled packages")
	log.Successln("build complete")
	log.Warnln("cache unavailable")
	log.Errorln("upload failed")
}
