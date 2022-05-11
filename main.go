package main

import (
	"bufio"
	"fmt"
	"github.com/gtjamesa/smtp-user-enum-go/enum"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func makeConnection(c *cli.Context) {
	fmt.Println("Got target", c.Args().Get(0))
	//enum.Execute(c.Args().Get(0))
	ReadFile(c.String("wordlist"))
}

func ReadFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

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

			enum.Execute(c)
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
