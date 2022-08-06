package main

import (
	"bufio"
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	s3lib "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Element struct {
	ObjectKey string `json:"objectkey"`
	Href      string `json:"href"`
	Expires   int    `json:"expires"`
}

var awsCfg aws.Config
var dynamo *dynamodb.Client
var s3 *s3lib.Client

var tableName string
var bucketName string

func main() {

	addr := flag.String("addr", "127.0.0.1:3000", "listening addr")
	flag.Parse()

	if err := initAws(); err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", middle(handle))

	log.Printf("Server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalln(err)
	}
}

func initAws() error {
	log.Println("Initializing DynamoDB client")
	var err error

	f, err := os.Open("vars")
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	tableName = scanner.Text()
	scanner.Scan()
	bucketName = scanner.Text()

	log.Printf("Read table name: %s\n", tableName)
	log.Printf("Read bucket name: %s\n", bucketName)

	awsCfg, err = config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "eu-central-1"
		return nil
	})
	if err != nil {
		return err
	}

	dynamo, err = dynamodb.NewFromConfig(awsCfg), nil
	if err != nil {
		return err
	}
	log.Println("DynamoDB client initialized successfully")

	s3 = s3lib.NewFromConfig(awsCfg)

	return err
}

func scanTable() ([]Element, error) {
	var items []Element

	out, err := dynamo.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		return items, err
	}

	err = attributevalue.UnmarshalListOfMaps(out.Items, &items)
	return items, err
}

func middle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		log.Printf("[ IN] %s\n", uri)
		next.ServeHTTP(w, r)
		log.Printf("[OUT] %s\n", uri)
	}

}

func handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Trying to scan table")
	items, err := scanTable()
	if err != nil {
		log.Printf("Could not scan table:\n%v\n", err)
	} else {
		log.Printf("Scanned table successfully:\n%v\n", items)
	}

	for idx, item := range items {
		log.Printf("Trying to get pre-signed URL for key '%s'\n", item.ObjectKey)
		psClient := s3lib.NewPresignClient(s3)
		sigV4Url, err := psClient.PresignGetObject(context.TODO(), &s3lib.GetObjectInput{
			Bucket: &bucketName,
			Key:    &item.ObjectKey,
		})
		if err != nil {
			log.Printf("Error when typing to sign url for key '%s'\t%v\t", item.ObjectKey, err)
		} else {
			log.Printf("Setting the item's URL to %s\n", sigV4Url.URL)
			items[idx].Href = sigV4Url.URL
		}

	}

	t, err := template.ParseFiles("./site.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	err = t.ExecuteTemplate(w, "main", items)
	if err != nil {
		log.Println(err)
	}
}
