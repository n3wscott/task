package main

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
	"log"
)

func main() {
	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	r := Receiver{client: client}
	if err := envconfig.Process("", &r); err != nil {
		panic(err)
	}

	if err := client.StartReceiver(context.Background(), r.Receive); err != nil {
		log.Fatal(err)
	}
}

type Receiver struct {
	client cloudevents.Client
	Target string `envconfig:"TARGET" required:"true"`
}

func (r *Receiver) Receive(event cloudevents.Event) {
	ctx := cloudevents.ContextWithTarget(context.Background(), r.Target)
	if _, _, err := r.client.Send(ctx, event); err != nil {
		fmt.Println(err)
	}
}
