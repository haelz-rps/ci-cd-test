package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kit/log/level"
	"github.com/ubix/dripper/manager/configs"
	"github.com/ubix/dripper/manager/ingestions"
	"github.com/ubix/dripper/manager/logging"
	hivemetastore "github.com/ubix/dripper/manager/repositories/data-catalog/hive-metastore"
	"github.com/ubix/dripper/manager/repositories/data-extractor/airbyte"
	"github.com/ubix/dripper/manager/repositories/kvstore/redis"

	// "github.com/ubix/dripper/manager/repositories/kvstore/inmem"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initiate Logger
	logger := config.GetLogger()
	brokerLogger := logging.NewGoKitWatermillLogger(logger)

	// Init repositories
	dataExtractor, err := airbyte.NewAirbyte(config.AirbyteConfig)
	if err != nil {
		level.Error(logger).Log("msg", err)
		panic(err)
	}

	dataCatalog, err := hivemetastore.NewHiveMetastore(config.MetastoreHost, config.MetastorePort, config.SparkHost, config.SparkPort, config.DataspaceHost, config.DataserviceHost)
	if err != nil {
		level.Error(logger).Log("msg", err)
		panic(err)
	}

	// store := inmem.NewInmemStore()
	store, err := redis.NewRedisStore(config.RedisURI)
	if err != nil {
		level.Error(logger).Log("msg", err)
		panic(err)
	}

	pub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{config.BrokerHost},
			Marshaler: kafka.DefaultMarshaler{},
		},
		brokerLogger,
	)
	if err != nil {
		level.Error(logger).Log("msg", err)
		panic(err)
	}

	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	// equivalent of auto.offset.reset: earliest
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	sub, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{config.BrokerHost},
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: saramaSubscriberConfig,
			ConsumerGroup:         config.KafkaGroupID,
		},
		brokerLogger,
	)
	if err != nil {
		level.Error(logger).Log("msg", err)
		panic(err)
	}
	// Init Services
	var connectorSvc ingestions.ConnectorService
	{
		connectorSvc = ingestions.NewConnectorService(dataExtractor, dataCatalog, pub, store)
		connectorSvc = ingestions.NewLoggingMiddleware(logger, connectorSvc)
	}

	// Init Endpoints
	var (
		connectorEndpointSet = ingestions.NewEndpointSet(connectorSvc)
	)

	r, err := message.NewRouter(message.RouterConfig{}, brokerLogger)
	if err != nil {
		panic(err)
	}
	ingestions.NewKafkaWorkerHandler(connectorEndpointSet, sub, pub, r)

	// Initiate kafka clients
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return r.Run(gCtx)
	})

	// Block main execution to wait for go rotines to return
	err = g.Wait()
	if err != nil {
		level.Error(logger).Log("msg", err)
	}
}
