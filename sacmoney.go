package sacmoney

import (
	cli "sacdev/sacmoney/pkg/cli"
	server "sacdev/sacmoney/pkg/server"
)

func DoCli() {
	cli.Run()
}

func RunServer() {
	server.Run()
}
