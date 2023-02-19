package main

import (
	"flag"
	"fmt"
	"log"
	"odo24_mobile_backend/api"
	"odo24_mobile_backend/config"
	"odo24_mobile_backend/db"
	"odo24_mobile_backend/sendmail"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	flag.Parse()

	log.Println("Start...")
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	// чтение настроек
	options := config.ReadConfig()

	fmt.Print("Create DB connection... ")
	err := db.CreateConnection(db.Options{
		DriverName:       options.Db.DriverName,
		ConnectionString: options.Db.ConnectionString,
		MaxIdleConns:     options.Db.MaxIdleConns,
		MaxOpenConns:     options.Db.MaxOpenConns,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("OK!")

	// ключ шифрования сессии
	//sessions.SetSecretKey(options.App.SessionKey)

	sendmail.InitSendmail()

	// инициализация API методов
	r := api.InitHandlers()
	fmt.Printf("Addr: %s\r\n", options.App.ServerAddr)
	err = r.Run(options.App.ServerAddr)
	if err != nil {
		panic(err)
	}
}
