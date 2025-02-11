package app

import "github.com/PIGcanstudy/gorder/payment/app/command"

type Application struct {
	Commands Commands
}

type Commands struct {
	CreatePayment command.CreatePaymentHandler
}
