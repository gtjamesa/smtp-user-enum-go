package enum

import (
	"bufio"
	"fmt"
	"github.com/gtjamesa/smtp-user-enum-go/enum/client"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SmtpEnum struct {
	ctx        *cli.Context
	resultChan chan string
}

func NewSmtpEnum(ctx *cli.Context) *SmtpEnum {
	return &SmtpEnum{
		ctx:        ctx,
		resultChan: make(chan string),
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

			fmt.Printf(resMsg)
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
	smtpClient := client.NewSmtpClient(s.ctx.Args().Get(0), s.ctx.Int("port"))

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

			//fmt.Printf("Got username: %s\n", username)
			succ, err := smtpClient.Vrfy(username)
			if err != nil {
				log.Fatal(err)
			}

			if succ {
				msg := fmt.Sprintf("User %s found\n", username)
				s.resultChan <- msg
			}
		}
	}
}

func (s *SmtpEnum) Run() {
	defer close(s.resultChan)

	var wg sync.WaitGroup
	threads := s.ctx.Int("threads")

	// We need to wait for all threads to finish
	wg.Add(threads)

	// Create buffered channel containing the wordlist
	wordChan := make(chan string, threads)

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
			close(wordChan)
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

// waitForExit will return a read-only channel and will only return once `runC` is closed
// the function listens for an interrupt signal and will allow a safe exit
// called from main code as "<-waitForExit()"
func waitForExit() <-chan struct{} {
	runC := make(chan struct{}, 1)
	sc := make(chan os.Signal, 1)

	// Listen to interrupt signal (Ctrl+C, SIGTERM, SIGQUIT, etc.)
	// When we receive a signal, it is fed into the `sc` channel
	signal.Notify(sc, os.Interrupt)

	go func() {
		// Close `runC` after the function has completed
		defer close(runC)
		// Read data from `sc`, this will block until `sc` channel has data
		// This stops the function from completing until we receive an interrupt signal
		<-sc
	}()

	return runC
}

type Scheduler struct {
	workers   int
	msgC      chan struct{}
	signalC   chan os.Signal
	waitGroup sync.WaitGroup
}

func NewScheduler(workers, buffer int) *Scheduler {
	return &Scheduler{
		// Amount of workers
		workers: workers,
		// Channel to receive events
		msgC: make(chan struct{}, buffer),
		// Channel to receive signals
		signalC: make(chan os.Signal, 1),
	}
}

func (s *Scheduler) ListenForWork() {
	go func() { // 1. Listen for messages to process
		signal.Notify(s.signalC, syscall.SIGTERM)

		for {
			<-s.signalC
			s.msgC <- struct{}{} // 2. Send to processing channel
		}
	}()

	s.waitGroup.Add(s.workers)

	for i := 0; i < s.workers; i++ {
		i := i
		go func() {
			for {
				select {
				case _, open := <-s.msgC: // 3. Wait for messages to process
					// Channel closed, exiting
					if !open {
						fmt.Printf("%d closing\n", i+1)
						s.waitGroup.Done()
						return
					}

					fmt.Printf("%d<- Processing\n", i)
				}
			}
		}()
	}
}

func (s *Scheduler) Exit() {
	close(s.msgC)
	s.waitGroup.Wait()
}
