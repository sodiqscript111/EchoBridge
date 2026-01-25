package temporal

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
)

var Client client.Client

func InitClient() error {
	hostPort := os.Getenv("TEMPORAL_HOST")
	if hostPort == "" {
		hostPort = "localhost:7233"
	}

	c, err := client.Dial(client.Options{
		HostPort: hostPort,
	})
	if err != nil {
		return err
	}
	Client = c
	log.Println("✅ Temporal client connected")
	return nil
}

func GetClient() client.Client {
	return Client
}

func Close() {
	if Client != nil {
		Client.Close()
	}
}
