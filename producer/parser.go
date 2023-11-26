package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
	"time"
)

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records[1:]
}

func requestUrl(url string) *http.Response {
	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		fmt.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
	}

	return res
}

func parseResponse(res *http.Response, url string) User {

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	html, _ := doc.Html()
	firstNameLastName := doc.Find("h1.page-title__title").Text()

	fmt.Printf("firstNameLastName: %s\n", firstNameLastName)

	return User{
		Url:  url,
		Name: firstNameLastName,
		Html: html,
	}
}

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

func userToString(user User) string {
	b, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Failed to parse into json: %s", err)
	}

	return string(b)
}

func main() {
	conn := getRabbitConnection()
	ch := getRabbitChannel(conn)
	queue := getRabbitQueue(ch)
	urls := readCsvFile("./habr.csv")

	for i := 0; i < 2; i++ {
		personData := requestUrl(urls[i][0])

		if personData.StatusCode == 200 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			body := parseResponse(personData, urls[i][0])
			err := ch.PublishWithContext(ctx,
				"",         // exchange
				queue.Name, // routing key
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(userToString(body)),
				})
			failOnError(err, "Failed to publish a message")
			log.Printf(" [x] Sent data of %s\n", body.Name)

			cancel()
		}
		time.Sleep(5 * time.Second)
	}
}
