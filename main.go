package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)


func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}


func ClearScreen() {
    if runtime.GOOS == "windows" {
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    } else {
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }
}


func main() {

	ClearScreen()

	conn, err := amqp.Dial("amqp://guest:guest@gamch1k.v6.navy:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a send channel")
	defer ch.Close()

	chl, err := conn.Channel()
	failOnError(err, "Failed to open a listen channel")
	defer chl.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Listen queue: ")
	ql, _ := reader.ReadString('\n')

	listen, err := chl.QueueDeclare(
		ql, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a listen queue")

	msgs, err := chl.Consume(
		listen.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	messages := []string{}
	
	go func() {
		var forever chan struct{}

		for d := range msgs {
			messages = append(messages, string(d.Body))
			ClearScreen()
			for _, msg := range messages {
				fmt.Print(msg)
			}
			fmt.Print(">> ")
		}

		<-forever
	}()
	
	
	fmt.Print("Send queue: ")
	qs, _ := reader.ReadString('\n')

	fmt.Print("Nickname: ")
	nick, _ := reader.ReadString('\n')
	nick = strings.Replace(nick, "\n", "", -1)

	send, err := ch.QueueDeclare(
		qs, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a send queue")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ClearScreen()
	
	for {
		fmt.Print(">> ")
		body, _ := reader.ReadString('\n')
	
		err = ch.PublishWithContext(ctx,
		"",     // exchange
		send.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprintf("%s << %s", nick, body)),
		})
		failOnError(err, "Failed to publish a message")

		messages = append(messages, fmt.Sprintf(">> %s", body))
	}

}
