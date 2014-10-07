package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

const (
	// policyRequest is the string which the client sends to initiate the transaction.
	policyRequest = "<policy-file-request/>\000"
)

var (
	portFlag     = flag.Int("port", 843, "Port to listen on")
	fileFlag     = flag.String("file", "", "Policy file to serve")
	userFlag     = flag.String("user", "", "User for privilege dropping (default to current user)")
	maxSizeFlag  = flag.Int("maxsize", 10000, "Maximum size of policy file")
	staticFlag   = flag.Bool("static", false, "If set, policy file is never checked for updates")
	updateFlag   = flag.Duration("update", 1*time.Minute, "Policy file check update period")
	deadlineFlag = flag.Duration("deadline", 1*time.Second, "Maximum time a request should take")
)

type versionedPolicy struct {
	body    []byte
	version int64
}
type versionedPolicyChan chan *versionedPolicy

// readPolicy reads a new policy document from the io.Reader and
// returns a versionedPolicy if validation succeeds.  The path
// argument is used for logging only.
func readPolicy(r io.Reader, path string, maxSize int, version int64) (*versionedPolicy, error) {
	newPolicy := make([]byte, maxSize+1)
	n, err := r.Read(newPolicy)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read %s: %v", path, err)
	}
	if n > maxSize {
		return nil, fmt.Errorf("%s is too large to be a policy file (>%v bytes)", path, maxSize)
	}

	newPolicy = newPolicy[:n]
	if !bytes.Contains(newPolicy, []byte("cross-domain-policy")) {
		return nil, fmt.Errorf("%s is not a valid policy file", path)
	}

	return &versionedPolicy{newPolicy, version}, nil
}

// loadPolicy returns the policy file at the given pathname
func loadPolicy(path string) (*versionedPolicy, error) {
	inf, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open %s: %v", path, err)
	}
	defer inf.Close()

	return readPolicy(inf, path, *maxSizeFlag, time.Now().Unix())
}

// policyService maintains a cached version of the policy, updating
// the policy every time a value is received on updateChan.  Clients
// may retrieve the latest policy from policyChan.  If the policy is
// not readable initially, it logs the error as fatal, otherwise it
// logs the error and continues.
func policyService(updateChan <-chan time.Time,
	policyChan versionedPolicyChan,
	policyLoader func() (*versionedPolicy, error)) {

	policy, err := policyLoader()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-updateChan:
			newPolicy, err := policyLoader()
			if err != nil {
				log.Print(err)
				break
			}
			if !bytes.Equal(policy.body, newPolicy.body) {
				log.Printf("Policy change detected - new policy %v is %v bytes",
					newPolicy.version,
					len(newPolicy.body))
				policy = newPolicy
			}
		case policyChan <- policy:
			// do nothing
		}
	}
}

// servePolicy handles a single incoming client.  It waits for the
// policy request and then sends the latest policy file.
func servePolicy(c net.Conn, deadline time.Duration, policyChan versionedPolicyChan) {
	defer c.Close()

	// This operation should be very short.  Ignoring error here
	// because the net.pipe we use for testing returns an error if
	// SetDeadline is called.
	c.SetDeadline(time.Now().Add(deadline))

	// Wait for request
	reqBuf := make([]byte, len(policyRequest))
	n, err := c.Read(reqBuf)
	if err != nil && err != io.EOF {
		log.Print("Error reading request: ", err)
		return
	}
	reqBuf = reqBuf[:n]

	if !bytes.Equal(reqBuf, []byte(policyRequest)) {
		log.Print("Invalid request.")
		return
	}

	// Retrieve policy to send
	usePolicy := <-policyChan

	// Send file
	log.Print("Sending policy ", usePolicy.version, " to ", c.RemoteAddr())
	_, err = c.Write(usePolicy.body)
	if err != nil {
		log.Print("Error writing policy response: ", err)
	}
}

func main() {
	flag.Parse()
	if *fileFlag == "" {
		log.Fatal("Required argument -file not found.")
	}

	var runAsUser *user.User
	var err error
	if *userFlag != "" {
		if runAsUser, err = user.Lookup(*userFlag); err != nil {
			log.Fatal(err)
		}
	} else {
		if runAsUser, err = user.Current(); err != nil {
			log.Fatal(err)
		}
	}

	policyChan := make(versionedPolicyChan)

	var updateChan <-chan time.Time
	if *staticFlag {
		updateChan = make(chan time.Time)
	} else {
		updateChan = time.Tick(*updateFlag)
	}
	go policyService(updateChan, policyChan,
		func() (*versionedPolicy, error) {
			return loadPolicy(*fileFlag)
		})

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *portFlag})
	if err != nil {
		log.Fatal("Couldn't open port ", *portFlag, ": ", err)
	}
	defer listener.Close()

	if runAsUser != nil {
		gid, err := strconv.Atoi(runAsUser.Gid)
		if err != nil {
			log.Fatal(err)
		}
		if syscall.Getgid() != gid {
			if err := syscall.Setgid(gid); err != nil {
				log.Fatal(err)
			}
		}
		uid, err := strconv.Atoi(runAsUser.Uid)
		if err != nil {
			log.Fatal(err)
		}
		if syscall.Getuid() != uid {
			if err := syscall.Setuid(uid); err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Print("*** Starting dl-flashpolicyd")
	log.Print("Serving file ", *fileFlag)
	log.Print("Listening on ", *portFlag)
	log.Print("Running as user ", runAsUser.Username)
	log.Print("Maximum policy file size is ", *maxSizeFlag)
	if *staticFlag {
		log.Print("Policy file will not be checked for updates.")
	} else {
		log.Print("Policy file will be checked for updates every ", *updateFlag)
	}
	log.Print("Incoming connection time limited to ", *deadlineFlag)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Print("Error accepting new connection.")
			continue
		}

		go servePolicy(client, *deadlineFlag, policyChan)
	}
}
