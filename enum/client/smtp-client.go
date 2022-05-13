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

func (c *SmtpClient) Close() {
	fmt.Println("Closing SMTP socket")
	err := c.socket.Close()

	if err != nil {
		return
	}
}

func (c *SmtpClient) Write(data string) error {
	msg := fmt.Sprintf("%s\r\n", data)
	_, err := c.socket.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// WriteRead will Write `data` to server and return the response
func (c *SmtpClient) WriteRead(data string) (string, error) {
	err := c.Write(data)

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

// WriteCheck Write `data` to server, and check for 250 at the start of the response
func (c *SmtpClient) WriteCheck(data string) (bool, error) {
	reply, err := c.WriteRead(data)

	if err != nil {
		log.Fatal(err)
	}

	// Check that the response starts with 250 for success
	found := strings.HasPrefix(reply, "250")
	return found, nil
}

// SendMethod TODO: Add enum
func (c *SmtpClient) SendMethod(method string, username string) (bool, error) {
	switch method {
	case "VRFY":
		return c.Vrfy(username)
	case "EXPN":
		return c.Expn(username)
	case "RCPT":
		return c.Rcpt(username)
	default:
		log.Fatal("here be dragons")
		return false, nil
	}
}

func (c *SmtpClient) Vrfy(username string) (bool, error) {
	return c.WriteCheck(fmt.Sprintf("VRFY %s", username))
}

func (c *SmtpClient) Expn(username string) (bool, error) {
	return c.WriteCheck(fmt.Sprintf("EXPN %s", username))
}

func (c *SmtpClient) Rcpt(username string) (bool, error) {
	// TODO: Needs to send "MAIL FROM:fake@example.com" once at the start of this enumeration mode
	return c.WriteCheck(fmt.Sprintf("RCPT TO:%s", username))
}

//type Probe struct {
//	test    string
//	allowed bool
//}
//
//func testing(james map[string]*Probe) {
//	fmt.Println("testing 1")
//	fmt.Println(james)
//	fmt.Println(james["VRFY"].allowed)
//	fmt.Println("----------")
//}
//
//// Probe will test the enumeration methods against the target to determine which are allowed
//func (c *SmtpClient) Probe() map[string]*Probe {
//	//methods := []string{"VRFY", "EXPN", "RCPT"}
//	probes := make(map[string]*Probe)
//
//	//for _, method := range methods {
//	//	reply, err := c.WriteRead(probe.test)
//	//
//	//	if err != nil {
//	//		probe.allowed = false
//	//		continue
//	//	}
//	//
//	//	probe.allowed = !strings.HasPrefix(reply, "502")
//	//}
//
//	// TODO: Look into why I can't access Probe as reference (&Probe)
//	probes["VRFY"] = &Probe{test: "VRFY root"}
//	probes["EXPN"] = &Probe{test: "EXPN root"}
//	probes["RCPT"] = &Probe{test: "RCPT TO: root"}
//
//	for _, probe := range probes {
//		reply, err := c.WriteRead(probe.test)
//
//		if err != nil {
//			probe.allowed = false
//			continue
//		}
//
//		probe.allowed = !strings.HasPrefix(reply, "502")
//	}
//
//	testing(probes)
//
//	return probes
//}
