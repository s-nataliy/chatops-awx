package main

import (
	controller "chatops/controller"
)

func main() {
	go controller.ListenHTTP()
	controller.TelegramBot()
}
