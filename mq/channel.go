package mq

type Channel string

const (
	METRICS      = Channel("metrics")
	LOGIN_SIGNIN = Channel("login.signin")
)
