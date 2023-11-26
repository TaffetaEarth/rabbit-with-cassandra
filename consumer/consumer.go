package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func getRabbitQueue(ch *amqp.Channel) amqp.Queue {
	q, err := ch.QueueDeclare(
		"user_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	return q
}

func getRabbitChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	return ch
}

func getRabbitConnection() *amqp.Connection {
	conn, err := amqp.Dial("amqp://rmuser:rmpassword@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	return conn
}

func main() {
	cluster := gocql.NewCluster("localhost")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "rmuser",
		Password: "rmpassword",
	}
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	conn := getRabbitConnection()
	ch := getRabbitChannel(conn)
	queue := getRabbitQueue(ch)

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var user User
			err := json.Unmarshal(d.Body, &user)

			err = session.Query("INSERT INTO rabbit.users (url, name, html) VALUES (?, ?, ?);", user.Url, user.Name, user.Html).Exec()
			log.Printf("Received a message with data of %s", user.Name)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	<-forever
}
