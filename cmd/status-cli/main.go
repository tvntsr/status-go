package main

import (
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/status-go/api"
	"github.com/status-im/status-go/protocol"

	"github.com/urfave/cli/v2"
)

const LightFlag = "light"
const InteractiveFlag = "interactive"
const CountFlag = "count"
const NameFlag = "name"
const AddFlag = "add"
const PortFlag = "port"

const RetrieveInterval = 300 * time.Millisecond
const SendInterval = 1 * time.Second
const WaitingInterval = 5 * time.Second

var CommonFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    LightFlag,
		Aliases: []string{"l"},
		Usage:   "Enable light mode",
	},
}

var DmFlags = append([]cli.Flag{
	&cli.BoolFlag{
		Name:    InteractiveFlag,
		Aliases: []string{"i"},
		Usage:   "Use interactive mode",
	},
	&cli.IntFlag{
		Name:    CountFlag,
		Aliases: []string{"c"},
		Value:   1,
		Usage:   "How many messages to sent from each user",
	},
}, CommonFlags...)

var ServeFlags = append([]cli.Flag{
	&cli.StringFlag{
		Name:    NameFlag,
		Aliases: []string{"n"},
		Value:   "Alice",
		Usage:   "Name of the user",
	},
	&cli.StringFlag{
		Name:    AddFlag,
		Aliases: []string{"a"},
		Usage:   "Add a friend with the public key",
	},
	&cli.IntFlag{
		Name:    PortFlag,
		Aliases: []string{"p"},
		Value:   8545,
		Usage:   "HTTP Server port to listen on",
	},
}, CommonFlags...)

var logger *zap.SugaredLogger

type StatusCLI struct {
	name      string
	messenger *protocol.Messenger
	backend   *api.GethStatusBackend
	logger    *zap.SugaredLogger
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "dm",
				Aliases: []string{"d"},
				Usage:   "Send direct message",
				Flags:   DmFlags,
				Action: func(cCtx *cli.Context) error {
					return simulate(cCtx)
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Start a server to send and receive messages",
				Flags:   ServeFlags,
				Action: func(cCtx *cli.Context) error {
					return serve(cCtx)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
