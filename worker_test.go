package nrt_configs

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

var client *firestore.Client
var ctx = context.Background()

func TestName(t *testing.T) {
	conf := &firebase.Config{ProjectID: os.Getenv("PROJECT_ID")}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	fmt.Println("firebase app is initialized.")

	snapIter := client.Collection("ooll-whitelist").Snapshots(ctx)
	/*defer snapIter.Stop()

	for {
		snap, err := snapIter.Next()
		if err != nil {
			log.Fatalln(err)
		}
		doc, _ := snap.Documents.GetAll()
		for _, data := range doc {
			fmt.Printf("%v\n", data.Data())
		}
		fmt.Printf("change size: %d\n", len(snap.Changes))
		for _, diff := range snap.Changes {
			fmt.Printf("diff: %+v\n", diff)
		}
		fmt.Println("----------")
	}*/
	ctx, cancelFunc := context.WithCancel(context.Background())
	wrkr := NewRTWorker(ctx, snapIter)
	wrkr.OnChange(func(docs []*firestore.DocumentSnapshot) error {
		for _, data := range docs {
			fmt.Printf("%v\n", data.Data())
		}
		return nil
	})

	go wrkr.Listen()

	time.Sleep(10 * time.Second)
	cancelFunc()

}
