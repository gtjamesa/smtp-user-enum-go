package client

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type SmtpClient struct {
	host   string
	port   int
	socket net.Conn
}

func NewSmtpClient(host string, port int) *SmtpClient {
	client := SmtpClient{
		host: host,
		port: port,
	}

	client.Connect(host, port)

	return &client
}

func (c *SmtpClient) Connect(host string, port int) {
	//fmt.Printf("[SmtpClient] Connecting to %v:%v\n", host, port)

	con, err := net.Dial("tcp", fmt.Sprintf("%v:%v", host, port))

	if err != nil {
		log.Fatal(err)
	}

	// Read banner from server
	reply := make([]byte, 1024)
	_, err = con.Read(reply)

	// Error reading from server
	if err != nil {
		log.Fatal(err)
	}

	c.socket = con
}

func (c *SmtpClient) write(data string) error {
	msg := fmt.Sprintf("%s\r\n", data)
	_, err := c.socket.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// writeRead will write `data` to server and return the response
func (c *SmtpClient) writeRead(data string) (string, error) {
	err := c.write(data)

	if err != nil {
		log.Fatal(err)
	}

	// Read response from server
	reply := make([]byte, 1024)
	_, err = c.socket.Read(reply)

	// Error reading from server
	if err != nil {
		return "", err
	}

	return string(reply), nil
}

// writeCheck Write `data` to server, and check for 250 at the start of the response
func (c *SmtpClient) writeCheck(data string) (bool, error) {
	reply, err := c.writeRead(data)

	if err != nil {
		log.Fatal(err)
	}

	// Check that the response starts with 250 for success
	found := strings.HasPrefix(reply, "250")
	return found, nil
}

func (c *SmtpClient) Vrfy(username string) (bool, error) {
	return c.writeCheck(fmt.Sprintf("VRFY %s", username))
}

func (c *SmtpClient) Expn(username string) (bool, error) {
	return c.writeCheck(fmt.Sprintf("EXPN %s", username))
}

func (c *SmtpClient) Rcpt(username string) (bool, error) {
	// TODO: Needs to send "MAIL FROM:fake@example.com" once at the start of this enumeration mode
	return c.writeCheck(fmt.Sprintf("RCPT TO:%s", username))
}
