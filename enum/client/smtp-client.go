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

func (client *SmtpClient) Connect(host string, port int) {
	fmt.Println(fmt.Sprintf("[SmtpClient] Connecting to %v:%v", host, port))

	con, err := net.Dial("tcp", fmt.Sprintf("%v:%v", host, port))

	if err != nil {
		log.Fatal(err)
	}

	client.socket = con
}

func (client *SmtpClient) write(data string) error {
	msg := fmt.Sprintf("%s\r\n", data)
	_, err := client.socket.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// writeCheck Write `data` to server, and check for 250 at the start of the response
func (client *SmtpClient) writeCheck(data string) (bool, error) {
	err := client.write(data)

	if err != nil {
		log.Fatal(err)
	}

	// Read response from server
	reply := make([]byte, 1024)
	_, err = client.socket.Read(reply)

	// Error reading from server
	if err != nil {
		return false, err
	}

	// Check that the response starts with 250 for success
	found := strings.HasPrefix(string(reply), "250")
	return found, nil
}

func (client *SmtpClient) Vrfy(username string) (bool, error) {
	return client.writeCheck(fmt.Sprintf("VRFY %s", username))
}

func (client *SmtpClient) Expn(username string) (bool, error) {
	return client.writeCheck(fmt.Sprintf("EXPN %s", username))
}

func (client *SmtpClient) Rcpt(username string) (bool, error) {
	// TODO: Needs to send "MAIL FROM:fake@example.com" once at the start of this enumeration mode
	return client.writeCheck(fmt.Sprintf("RCPT TO:%s", username))
}
