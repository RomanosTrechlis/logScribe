package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"golang.org/x/net/context"

	pb "github.com/RomanosTrechlis/go-scribe/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	cert = "../../certs/client.crt"
	key  = "../../certs/client.key"
	ca   = "../../certs/CertAuth.crt"
)

func main() {
	var scribe string
	var filename string
	sec := flag.Bool("s", false, "true for secure connection")
	flag.StringVar(&scribe, "addr", ":8080", "scribe's address")
	flag.StringVar(&filename, "filename", "test", "filename to write the logs")
	flag.Parse()

	s := strings.Split(scribe, ":")
	server := s[0]
	port := ":" + s[1]
	if server == "" {
		server = "127.0.0.1"
	}

	var conn *grpc.ClientConn
	if *sec {
		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			fmt.Printf("could not load client key pair: %s", err)
			return
		}

		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(ca)
		if err != nil {
			fmt.Printf("could not read ca certificate: %s", err)
			return
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			fmt.Printf("failed to append ca certs")
			return
		}

		creds := credentials.NewTLS(&tls.Config{
			ServerName:   server, // NOTE: this is required!
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})

		// Create a connection with the TLS credentials
		conn, err = grpc.Dial(server+port,
			grpc.WithTransportCredentials(creds),
			grpc.WithTimeout(1*time.Second))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
	} else {
		var err error
		conn, err = grpc.Dial(server+port,
			grpc.WithInsecure(),
			grpc.WithTimeout(1*time.Second))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
	}

	var i int
	for {
		i++
		c := pb.NewLogScribeClient(conn)
		req := &pb.LogRequest{
			Filename: filename,
			Path:     "path",
			Line:     fmt.Sprintf("%d: This is a test", i),
		}
		r, err := c.Log(context.Background(), req)
		if err != nil {
			log.Fatalf("failled: %v", err)
		}
		fmt.Printf("%s", r.Res)
	}

}
