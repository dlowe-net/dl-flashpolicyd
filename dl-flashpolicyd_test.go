package main

import (
	"bytes"
	"io"
	"net"
	"runtime"
	"testing"
	"time"
)

func TestReadPolicy(t *testing.T) {
	cases := []struct {
		body    string
		path    string
		maxSize int
		wantErr bool
		version int64
	}{
		{`<cross-domain-policy></cross-domain-policy>`, "sample.txt", 10000, false, 872472},
		{`foo`, "sample.txt", 10000, true, 872472},
		{`<cross-domain-policy></cross-domain-policy>`, "sample.txt", 5, true, 872472},
	}

	for i, tt := range cases {
		var buf bytes.Buffer
		buf.Reset()
		buf.WriteString(tt.body)
		got, err := readPolicy(&buf, tt.path, tt.maxSize, tt.version)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%d. want error, got nil", i)
			}
			continue
		}
		if tt.body != string(got.body) {
			t.Errorf("%d. want %d, got %d", tt.body, string(got.body))
		}
		if tt.version != got.version {
			t.Errorf("%d. want %d, got %d", tt.version, got.version)
		}
	}
}

func TestPolicyService(t *testing.T) {
	updateChan := make(chan time.Time)
	policyChan := make(versionedPolicyChan)
	testBody := `<cross-domain-policy></cross-domain-policy>`
	testVersion := int64(1234)
	policyLoader := func() (*versionedPolicy, error) {
		return &versionedPolicy{[]byte(testBody), testVersion}, nil
	}
	go policyService(updateChan, policyChan, policyLoader)
	runtime.Gosched()

	got := <-policyChan
	if string(got.body) != testBody {
		t.Errorf("want %s, got %s", testBody, got.body)
	}
	if got.version != testVersion {
		t.Errorf("want %s, got %s", testVersion, got.version)
	}

	testBody = `<cross-domain-policy>foo</cross-domain-policy>`
	testVersion = 1235
	updateChan <- time.Now()
	got = <-policyChan
	if string(got.body) != testBody {
		t.Errorf("want %s, got %s", testBody, got.body)
	}
	if got.version != testVersion {
		t.Errorf("want %s, got %s", testVersion, got.version)
	}
}

func TestServePolicy(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	policyChan := make(versionedPolicyChan)

	testBody := `<cross-domain-policy>foo</cross-domain-policy>`
	testVersion := int64(1235)
	go func() {
		policyChan <- &versionedPolicy{[]byte(testBody), testVersion}
	}()

	go servePolicy(serverConn, 1*time.Minute, policyChan)
	n, err := clientConn.Write([]byte(policyRequest))
	if err != nil {
		t.Error(err)
	}
	if n != len(policyRequest) {
		t.Errorf("written: want %v, got %v", len(policyRequest), n)
	}
	response := make([]byte, 256)
	n, err = clientConn.Read(response)
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	response = response[:n]

	if string(response) != testBody {
		t.Errorf("want %v, got %v", testBody, string(response))
	}
}
