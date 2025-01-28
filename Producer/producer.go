package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
)

// Comment struct
type TIFDStatus struct {
	Text     string `json:"text"`
	AcctNo   int    `json:"acctno"`
	TxnType  int    `json:"txntype"`
	UniqueId int    `json:"uniqueid"`
}

func main() {

	app := fiber.New()
	api := app.Group("/api/v1") /*API Group*/

	api.Post("/TIFDSTATUS", createTIFDRequest)

	app.Listen(":3000")
}

func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func PushCommentToQueue(topic string, message []byte) error {

	brokersUrl := []string{"localhost:29092"}
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}

	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

	return nil
}

// createTIFDRequest handler
func createTIFDRequest(c *fiber.Ctx) error {

	// Instantiate new Message struct
	cmt := new(TIFDStatus)

	//  Parse body into comment struct
	if err := c.BodyParser(cmt); err != nil {
		log.Println(err)
		c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
		return err
	}
	// convert body into bytes and send it to kafka as a serailized ddata
	cmtInBytes, err := json.Marshal(cmt)
	if err != nil {
		log.Println("Error in conversion of body into bytes")
		return err
	}

	err = PushCommentToQueue("TIFDStatus", cmtInBytes)
	if err != nil {
		log.Println("Error in connecting to producer or pushing data to producer")
		return err
	}

	// Return Comment in JSON format
	err = c.JSON(&fiber.Map{
		"success": true,
		"message": "TIFD Request pushed successfully",
		"comment": cmt,
	})
	if err != nil {
		c.Status(500).JSON(&fiber.Map{
			"success": false,
			"message": "Error creating product",
		})
		return err
	}

	return err
}
