package nrt_configs

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestConfigmap(t *testing.T) {
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

	collection := client.Collection("ooll-whitelist")

	cm := NewConfigmap()
	t.Log("Register.error", cm.Register("ooll-whitelist", collection))
	cm.Listen()
	t.Log("before", cm.Exists("ooll-whitelist", "Element"))

	dr, _, _ := collection.Add(context.Background(), ConfigValue{
		Name:  "Element",
		Value: true,
	})
	time.Sleep(10 * time.Second)
	t.Log("after", cm.Exists("ooll-whitelist", "Element"))
	dr.Delete(context.Background())
}

func TestConfigmapConcurrent(t *testing.T) {
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

	collection := client.Collection("ooll-whitelist")

	cm := NewConfigmap()
	fmt.Println("Register.error", cm.Register("ooll-whitelist", collection))
	cm.Listen()

	fmt.Println("Element exists? before", cm.Exists("ooll-whitelist", "Element"))

	dr, _, _ := collection.Add(context.Background(), ConfigValue{
		Name:  "Element",
		Value: -1,
	})
	go func() {
		//Ticker
		for i := range make([]byte, 10) {
			time.Sleep(time.Second)
			dr.Set(context.Background(), ConfigValue{
				Name:  "Element",
				Value: i,
			})
		}
	}()

	for i := range make([]byte, 20) {
		fmt.Println("Value of Element after", i, "seconds", cm.Get("ooll-whitelist", "Element"))
		time.Sleep(time.Second)
	}

	dr.Delete(context.Background())
}
