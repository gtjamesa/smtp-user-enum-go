package main

import (
	"github.com/gtjamesa/smtp-user-enum-go/enum"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func main() {
	var wordlist string

	app := &cli.App{
		Name:        "SMTP User Enum",
		Usage:       "A simple SMTP user enumeration program",
		Description: "Multiple targets can be specified as arguments.",
		HelpName:    "smtp-user-enum",
		ArgsUsage:   "<targets>",
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.Exit("One or more targets must be specified", 1)
			}

			app := enum.NewSmtpEnum(c)
			app.Run()

			return nil
		},
		//Commands: []*cli.Command{
		//	{
		//		Name:  "ip",
		//		Usage: "Resolve FQDN to IPv4 address.",
		//		Action: func(c *cli.Context) error {
		//			fmt.Println(c.Args())
		//			ns, nsErr := net.LookupIP(c.String("host"))
		//			if nsErr != nil {
		//				return nsErr
		//			}
		//			for _, v := range ns {
		//				fmt.Println(v)
		//			}
		//			return nil
		//		},
		//	},
		//},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   25,
				Usage:   "Set a non-standard SMTP `port`",
			},
			&cli.StringFlag{
				Name:        "wordlist",
				Aliases:     []string{"w"},
				Usage:       "Wordlist containing usernames",
				Destination: &wordlist,
				Required:    true,
			},
			&cli.IntFlag{
				Name:    "threads",
				Aliases: []string{"t"},
				Usage:   "Amount of `threads` to use",
				Value:   8,
			},
			&cli.StringFlag{
				Name:    "method",
				Aliases: []string{"m"},
				Usage:   "Enumeration `method` to use (allowed: VRFY, EXPN, RCPT)",
				Value:   "VRFY",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Set verbose output",
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
