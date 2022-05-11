package enum

import (
	"fmt"
	"github.com/gtjamesa/smtp-user-enum-go/enum/client"
	"github.com/urfave/cli/v2"
	"log"
)

func connect(target string, port int) {
	var err error
	var succ bool

	fmt.Println(fmt.Sprintf("Connecting to %v:%v", target, port))

	//con, err := net.Dial("tcp", fmt.Sprintf("%v:%v", target, port))
	//checkErr(err)
	//defer con.Close()

	smtpClient := client.NewSmtpClient(target, port)

	succ, err = smtpClient.Vrfy("root")
	checkErr(err)
	if succ {
		fmt.Println("User root found")
	}
	succ, err = smtpClient.Vrfy("james")
	checkErr(err)
	if succ {
		fmt.Println("User james found")
	}
	//smtpClient.Connect(target, port)
	//&smtpClient.Vrfy("james")

	//return con

	//msg := "This is a test\n"
	//
	//_, err = con.Write([]byte(msg))
	//checkErr(err)
	//
	//reply := make([]byte, 1024)
	//_, err = con.Read(reply)
	//checkErr(err)
	//
	//fmt.Println(string(reply))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Execute(c *cli.Context) {
	var target string = c.Args().Get(0)

	connect(target, c.Int("port"))
}
