package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	tls_client "github.com/albinstman/tls-client"
	"github.com/albinstman/tls-client/profiles"
	tls "github.com/albinstman/utls"
	http "github.com/bogdanfinn/fhttp"
	quic "github.com/bogdanfinn/quic-go-utls"
)

func main() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_131),
		tls_client.WithProtocolRacing(),
		tls_client.WithHTTP3Dial(func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
			// Example: dial QUIC over a specific local UDP address.
			// You could also route through a tunnel, custom DNS, etc.
			log.Printf("[http3-dial] dialing QUIC to %s", addr)

			udpAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				return nil, fmt.Errorf("resolve udp addr: %w", err)
			}

			udpConn, err := net.ListenUDP("udp", nil)
			if err != nil {
				return nil, fmt.Errorf("listen udp: %w", err)
			}

			conn, err := quic.Dial(ctx, udpConn, udpAddr, tlsCfg, cfg)
			if err != nil {
				udpConn.Close()
				return nil, fmt.Errorf("quic dial: %w", err)
			}

			log.Printf("[http3-dial] connected to %s via %s", addr, udpConn.LocalAddr())
			return conn, nil
		}),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}

	req.Header = http.Header{
		"accept":     {"*/*"},
		"user-agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("read body: %v", err)
	}

	log.Printf("Status: %d", resp.StatusCode)
	log.Printf("Response:\n%s", string(body))
}
