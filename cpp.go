package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/rs/xid"
	"github.com/streadway/amqp"
)

// Request ...
type Request struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	Filename    string    `json:"-"`
	Outfile     string    `json:"-"`
	Output      string    `json:"output"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
	Status      int       `json:"status"`
}

func generateUniqueID() string {
	return xid.New().String()
}

func updateStatus(r *Request) error {

	jsonBody, _ := json.Marshal(r)

	return queueConf.channel.Publish(
		"",
		"DB_UPDATE",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/json",
			Body:        []byte(jsonBody),
		},
	)

}

func createFile(r *Request) error {

	r.Filename = fmt.Sprintf("%s.cpp", r.ID)

	f, err := os.Create(r.Filename)

	if err != nil {
		return err
	}

	l, err := f.WriteString(r.Code)

	if err != nil {
		f.Close()
		os.Remove(r.Filename)
		return err
	}

	log.Printf("Successfully written: %s %d bytes\n", r.Filename, l)

	f.Close()

	return nil
}

func compileCode(r *Request) error {

	r.Outfile = fmt.Sprintf("%s.out", r.ID)

	cmd := exec.Command("g++", r.Filename, "-o", r.Outfile)

	err := cmd.Run()

	if err != nil {
		return err
	}

	log.Println("Successfully compiled", r.Filename)

	return nil
}

func captureOutput(r *Request) error {

	log.Println("Running", r.Outfile)

	r.StartedAt = time.Now().UTC()

	cmd := exec.Command(fmt.Sprintf("./%s", r.Outfile))

	out, err := cmd.CombinedOutput()

	r.CompletedAt = time.Now().UTC()

	if err != nil {
		return err
	}

	r.Output = string(out)
	return nil
}

func processCpp(request *Request) error {

	err := createFile(request)

	if err != nil {
		return err
	}

	err = compileCode(request)

	if err != nil {
		return err
	}

	request.Status = 1

	err = updateStatus(request)

	if err != nil {
		return err
	}

	err = captureOutput(request)

	if err != nil {
		return err
	}

	request.Status = 2

	err = updateStatus(request)

	if err != nil {
		return err
	}

	os.Remove(request.Filename)
	os.Remove(request.Outfile)

	return nil
}

// HandleCpp ...
func HandleCpp(r *Request) error {

	err := processCpp(r)

	if err != nil {
		return err
	}

	log.Println(r.Output)

	return nil
}
