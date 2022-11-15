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
	datasetID = "test_dataset"
	tableID   = "test_table"
	timeout   = 300
)

var (
	// For information about nested schema, see below
	// https://cloud.google.com/bigquery/docs/samples/bigquery-nested-repeated-schema
	schema = bigquery.Schema{
		{
			Name:     "labels",
			Required: false,
			Type:     bigquery.RecordFieldType,
			Schema: bigquery.Schema{
				{Name: "log_id", Required: false, Type: bigquery.StringFieldType},
			},
		},
	}
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
	// The following option imitates the one of Spanner emulator.
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

	// Creating a dataset and a table
	dataset := cli.Dataset(datasetID)
	defer clean(dataset)
	if err := dataset.Create(ctx, nil); err != nil {
		return errors.Wrap(err, "error Create dataset")
	}
	dsMeta, err := dataset.Metadata(ctx)
	if err != nil {
		return errors.Wrap(err, "error dataset.Metadata")
	}
	log.Printf("[DEBUG] dataset created: %#v\n", dsMeta)
	table := dataset.Table(tableID)
	if err := table.Create(ctx, &bigquery.TableMetadata{
		Schema: schema,
	}); err != nil {
		return errors.Wrap(err, "error Create table")
	}
	tbMeta, err := table.Metadata(ctx)
	if err != nil {
		return errors.Wrap(err, "error table.Metadata")
	}
	log.Printf("[DEBUG] table created: %#v\n", tbMeta)
	return nil
}

func clean(dataset *bigquery.Dataset) {
	if err := dataset.DeleteWithContents(context.TODO()); err != nil {
		log.Printf("[ERROR] error dataset.DeleteWithContents: %s\n", err)
		return
	}
	log.Printf("[DEBUG] dataset is deleted")
}
