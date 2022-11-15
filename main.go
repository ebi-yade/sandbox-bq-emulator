package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	projectID = "test-project"
	timeout   = 300
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("[INFO] process successfully finished")
}

func run(ctx context.Context) error {
	addr := os.Getenv("BIGQUERY_EMULATOR_HOST")
	if addr == "" {
		return errors.New("error BIGQUERY_EMULATOR_HOST must not be empty")
	}
	// 参考: Spanner SDK がエミュレータのアドレス切り替えに使っている部分
	// https://github.com/googleapis/google-cloud-go/blob/5b307584fcd635635aae6a2fc4ba8252f2bbe22d/spanner/client.go#L177-L186
	opts := []option.ClientOption{
		option.WithEndpoint(addr),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}
	cli, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		return errors.Wrap(err, "error NewClient")
	}
	log.Printf("[DEBUG] client created: %#v\n", cli)
	return nil
}
