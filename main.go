package main

import (
	"context"
	"log"
	"man-counter/consumer"
	"man-counter/devices"
	"man-counter/downstream"
	"man-counter/storage"
	"man-counter/visitors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	UpdateIntervalEnvName = "UPDATE_INTERVAL"

	PostgresUrlEnvName  = "POSTGRES_URL"
	PostgresUserEnvMane = "POSTGRES_USER"
	PostgresPassEnvName = "POSTGRES_PASS"
	PostgresDBEnvName   = "POSTGRES_DB"

	KafkaUrlEnvName     = "KAFKA_URL"
	KafkaGroupIdEnvName = "KAFKA_GROUP_ID"
	OutQueueEnvName     = "OUT_QUEUE_NAME"
	CamQueueEnvName     = "CAM_QUEUE_NAME"
	DeviceQueueEnvName  = "DEVICE_QUEUE_NAME"

	defaultSleepTime = 2
)

var (
	outQueueName    = os.Getenv(OutQueueEnvName)
	camQueueName    = os.Getenv(CamQueueEnvName)
	deviceQueueName = os.Getenv(DeviceQueueEnvName)
	updateInterval  = os.Getenv(UpdateIntervalEnvName)
	postgresUrl     = os.Getenv(PostgresUrlEnvName)
	postgresUser    = os.Getenv(PostgresUserEnvMane)
	postgresPass    = os.Getenv(PostgresPassEnvName)
	postgresDb      = os.Getenv(PostgresDBEnvName)
	kafkaUrl        = os.Getenv(KafkaUrlEnvName)
	kafkaGroupId    = os.Getenv(KafkaGroupIdEnvName)
)

func main() {

	sleepTime, err := strconv.Atoi(updateInterval)
	if err != nil {
		sleepTime = defaultSleepTime
	}

	st, err := storage.InitStorage(postgresDb, postgresUser, postgresPass, postgresUrl)
	if err != nil {
		log.Fatalln("Error initing db:", err.Error())
	}
	defer st.Close()

	ctx := context.Background()
	errorChannel := make(chan error)

	visitorConsumer := consumer.New(ctx, kafkaUrl, kafkaGroupId, camQueueName)
	visitorHandler := visitors.New(st)

	activeDevicesConsumer := consumer.New(ctx, kafkaUrl, kafkaGroupId, deviceQueueName)
	activeDeviceHandler := devices.New(st)

	downstreamSender := downstream.New(ctx, st, kafkaUrl, outQueueName, sleepTime)

	visitorConsumer.StartConsumeAsync(visitorHandler, errorChannel)
	activeDevicesConsumer.StartConsumeAsync(activeDeviceHandler, errorChannel)
	downstreamSender.StartNotifyAsync(errorChannel)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	select {
	case e := <-errorChannel:
		log.Println("Got error from one of executing gorutines", e.Error())
	case <-signals:
		log.Println("Got signal...")
	}

	downstreamSender.StopNotify()
	visitorConsumer.StopConsume()
	activeDevicesConsumer.StopConsume()

	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()
}
