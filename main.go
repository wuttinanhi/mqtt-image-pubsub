package main

import (
	"fmt"
	"io"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MQTT_SERVER = "tcp://localhost:1883"

var MQTT_TOPIC = "image_topic"

func createClient(clientID string) mqtt.Client {
	// Create an MQTT client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(MQTT_SERVER)

	opts.SetClientID(clientID)
	opts.SetUsername("rabbit")
	opts.SetPassword("rabbit")

	// FOR DEBUGGING
	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Error connecting to MQTT server: %v\n", token.Error())
		os.Exit(1)
	}

	fmt.Println("Connected to MQTT server")

	return client
}

func Producer(client mqtt.Client, topic string) {
	fmt.Println("publishing to topic: ", topic)

	// Read the image file
	imageFile, err := os.Open("./test.jpg")
	if err != nil {
		fmt.Printf("Error opening image file: %v\n", err)
		os.Exit(1)
	}
	defer imageFile.Close()

	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		os.Exit(1)
	}

	// Publish the image to the MQTT topic
	token := client.Publish(topic, 2, false, imageData)
	token.Wait()

	if token.Error() != nil {
		fmt.Printf("Error publishing image: %v\n", token.Error())
	} else {
		fmt.Printf("image sent to topic: %s\n", topic)
	}

	fmt.Println()
}

func messageHandler(client mqtt.Client, msg mqtt.Message) {
	messageID := msg.MessageID()

	// Handle incoming MQTT messages here
	fmt.Printf("Received message\ntopic: %s\nmessage: %v\n", msg.Topic(), messageID)

	imageData := msg.Payload()

	// Save the image to a file
	err := os.WriteFile("./received.jpg", imageData, 0644)
	if err != nil {
		fmt.Printf("Error saving received image: %v\n", err)
		return
	}

	// ack
	msg.Ack()

	fmt.Println("Received image saved as 'received.jpg'")
	fmt.Println()
}

func Subscribe(client mqtt.Client, topic string) {
	fmt.Println("listening to topic: ", topic)

	token := client.Subscribe(topic, 2, messageHandler)

	// Wait for the subscription to complete
	token.Wait()

	if token.Error() != nil {
		fmt.Printf("Error subscribing to topic: %v\n", token.Error())
		os.Exit(1)
	}

	<-make(chan struct{})
}

func main() {
	mode := os.Args[1]

	if mode == "sub" {
		client := createClient("go-sub")
		Subscribe(client, MQTT_TOPIC)
	} else {
		client := createClient("go-pub")
		Producer(client, MQTT_TOPIC)
	}
}
