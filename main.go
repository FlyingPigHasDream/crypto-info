package main

import (
	cryptov1 "crypto-info/kitex_gen/crypto/v1/healthservice"
	"log"
)

func main() {
	svr := cryptov1.NewServer(new(HealthServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
