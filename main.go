package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/knqyf263/utern/cloudwatch"
	"github.com/knqyf263/utern/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var (
	version = "dev"
	commit  string
)

func main() {
	app := cli.NewApp()
	app.Name = "utern"
	app.Usage = "Multi group and stream log tailing for AWS CloudWatch Logs"
	app.Version = fmt.Sprintf("%s (rev:%s)", version, commit)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "stream, n",
			Value: "",
			Usage: "Log stream name (regular expression).\n\tDisplays all if omitted. If the option\n\t\"since\" is set to recent time, this option\n\tusually makes it faster than the option\n\t\"stream-prefix\"",
		},
		cli.StringFlag{
			Name:  "stream-prefix, p",
			Value: "",
			Usage: "Log stream name prefix. If a log group\n\tcontains many log streams, this option makes\n\tit faster.",
		},
		cli.StringFlag{
			Name:  "since, s",
			Value: "5m",
			Usage: "Return logs newer than a relative duration\n\tlike 52, 2m, or 3h.",
		},
		cli.StringFlag{
			Name:  "end, e",
			Value: "",
			Usage: "Return logs older than a relative duration\n\tlike 0, 2m, or 3h.",
		},
		cli.StringFlag{
			Name:  "profile",
			Value: "",
			Usage: "Specify an AWS profile.",
		},
		cli.StringFlag{
			Name:  "code",
			Value: "",
			Usage: "Specify MFA token code directly\n\t(if applicable), instead of using stdin.",
		},
		cli.StringFlag{
			Name:  "region, r",
			Value: "",
			Usage: "Specify an AWS region.",
		},
		cli.StringFlag{
			Name:  "filter",
			Value: "",
			Usage: "The filter pattern to use. For more\n\tinformation, see https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html.",
		},
		cli.BoolFlag{
			Name:  "timestamps",
			Usage: "Print timestamps",
		},
		cli.BoolFlag{
			Name:  "event-id",
			Usage: "Print event ID",
		},
		cli.BoolFlag{
			Name:  "no-log-group",
			Usage: "Suppress display of log group name",
		},
		cli.BoolFlag{
			Name:  "no-log-stream",
			Usage: "Suppress display of log stream name",
		},
		cli.IntFlag{
			Name:  "max-length",
			Value: 0,
			Usage: "Maximum log message length",
		},
		cli.BoolFlag{
			Name:  "color",
			Usage: "Force color output even if not a tty",
		},
	}

	app.Action = func(c *cli.Context) error {
		config, err := config.New(c)
		if err != nil {
			return errors.Wrap(err, "option error")
		}
		return run(config)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func handleSignal(cancel context.CancelFunc) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		for {
			s := <-signalChannel
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				cancel()
			}
		}
	}()
}

func run(conf *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	handleSignal(cancel)
	client := cloudwatch.NewClient(conf)
	return client.Tail(ctx)
}
