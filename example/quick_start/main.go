package main

import "github.com/cuteLittleDevil/go-jt808/service"

func main() {
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:8080"),
		service.WithNetwork("tcp"),
	)
	goJt808.Run()
}
