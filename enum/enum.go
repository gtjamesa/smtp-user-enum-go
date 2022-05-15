package enum

import (
	"bufio"
	"fmt"
	"github.com/gtjamesa/smtp-user-enum-go/enum/client"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type Probe struct {
	test    string
	allowed bool
}

type Result struct {
	username string
	reply    string
}

type SmtpEnum struct {
	ctx        *cli.Context
	targets    []string
	method     string
	resultChan chan Result
}

func NewSmtpEnum(ctx *cli.Context) *SmtpEnum {
	return &SmtpEnum{
		ctx:        ctx,
		targets:    ctx.Args().Slice(),
		method:     ctx.String("method"),
		resultChan: make(chan Result),
	}
}

// getWordlist will return a file stream iterator
func (s *SmtpEnum) getWordlist(filePath string) (*bufio.Scanner, error) {
	// Read wordlist from stdin
	if filePath == "-" {
		return bufio.NewScanner(os.Stdin), nil
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open wordlist: %s", err)
	}

	return bufio.NewScanner(file), nil
}

func (s *SmtpEnum) showResults() {
	for {
		select {
		case resMsg, ok := <-s.resultChan:
			if !ok {
				return
			}

			if s.ctx.Bool("verbose") {
				fmt.Printf("%s\t\t%s", resMsg.username, resMsg.reply)
				continue
			}

			fmt.Printf("%s\n", resMsg.username)
		}
	}
}

// worker is responsible for sending data to the SMTP server
// Each worker will have its own connection to the target
// wordChan is a read-only channel containing words from the wordlist
// wg is a WaitGroup pointer to the main WaitGroup
func (s *SmtpEnum) worker(wordChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Open connection to target SMTP server
	smtpClient := client.NewSmtpClient(s.targets[0], s.ctx.Int("port"))

	for {
		select {
		case <-s.ctx.Done():
			return
		// Read username from `wordChan`
		case username, ok := <-wordChan:
			// `wordChan` has been closed
			if !ok {
				return
			}

			succ, reply, err := smtpClient.SendMethod(s.method, username)
			if err != nil {
				log.Fatal(err)
			}

			if succ {
				s.resultChan <- Result{
					username: username,
					reply:    reply,
				}
			}
		}
	}
}

// Probe will test the enumeration methods against the target to determine which are allowed
func (s *SmtpEnum) Probe() (map[string]*Probe, error) {
	methods := []string{"VRFY", "EXPN", "RCPT"}
	probes := make(map[string]*Probe)

	smtpClient := client.NewSmtpClient(s.targets[0], s.ctx.Int("port"))
	defer smtpClient.Close()

	// Define probes to be sent to the target
	probes["VRFY"] = &Probe{test: "VRFY root"}
	probes["EXPN"] = &Probe{test: "EXPN root"}
	probes["RCPT"] = &Probe{test: "RCPT TO: root"}

	fmt.Printf(smtpClient.GetBanner())

	for _, probe := range probes {
		// Send the Probe
		reply, err := smtpClient.WriteRead(probe.test)

		if err != nil {
			probe.allowed = false
			continue
		}

		// The Probe is disallowed if we receive a 502 response
		probe.allowed = !strings.HasPrefix(reply, "502")
	}

	// Check that the user-defined method is allowed
	if probes[s.method].allowed {
		return probes, nil
	}

	fmt.Printf("%s method disallowed by server\n", s.method)

	// Switch to first available method
	var found = false
	for _, method := range methods {
		if probes[method].allowed {
			s.method = method
			found = true
			break
		}
	}

	if !found {
		return probes, fmt.Errorf("no available enumeration methods found")
	}

	return probes, nil
}

func (s *SmtpEnum) Run() {
	defer close(s.resultChan)

	var wg sync.WaitGroup
	threads := s.ctx.Int("threads")

	// We need to wait for all threads to finish
	wg.Add(threads)

	// Create buffered channel containing the wordlist
	wordChan := make(chan string, threads)

	// Listen for interrupt calls to exit cleanly
	listenForInterrupt()

	// Probe the server for available enumeration methods
	if _, err := s.Probe(); err != nil {
		log.Fatal(err)
	}

	// Start `threads` workers that are listening to the wordlist channel
	// The channel will be populated with words from the wordlist file
	for i := 0; i < threads; i++ {
		go s.worker(wordChan, &wg)
	}

	// Read the wordlist and return a scanner
	scanner, err := s.getWordlist(s.ctx.String("wordlist"))
	if err != nil {
		log.Fatal(err)
	}

	// Print results to the screen as they come in from the workers
	go s.showResults()

	// Iterate the wordlist scanner
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Read a word from the wordlist scanner
			word := scanner.Text()
			// Add the word to the wordlist channel
			wordChan <- word
		}
	}

	close(wordChan)
	wg.Wait()
}

func listenForInterrupt() {
	// Signal channel
	sc := make(chan os.Signal, 1)

	go func() {
		// Listen to interrupt signal (Ctrl+C, SIGTERM, SIGQUIT, etc.)
		// When we receive a signal, it is fed into the `sc` channel
		signal.Notify(sc, os.Interrupt)
		<-sc
		os.Exit(0)
	}()
}
