package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	// "github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log/level"
	"github.com/ubix/dripper/manager/configs"
	"github.com/ubix/dripper/manager/docs"
	"github.com/ubix/dripper/manager/ingestions"
	"github.com/ubix/dripper/manager/logging"
	hivemetastore "github.com/ubix/dripper/manager/repositories/data-catalog/hive-metastore"
	"github.com/ubix/dripper/manager/repositories/data-extractor/airbyte"
	"github.com/ubix/dripper/manager/repositories/kvstore/redis"

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
	// Initiate Repositories
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

	// saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	// // equivalent of auto.offset.reset: earliest
	// saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	// sub, err := kafka.NewSubscriber(
	// 	kafka.SubscriberConfig{
	// 		Brokers:               []string{brokerHost},
	// 		Unmarshaler:           kafka.DefaultMarshaler{},
	// 		OverwriteSaramaConfig: saramaSubscriberConfig,
	// 	},
	// 	brokerLogger,
	// )
	// if err != nil {
	// 	level.Error(logger).Log("msg", err)
	// 	panic(err)
	// }
	// Initiate Services
	ingestionsService := ingestions.NewConnectorService(dataExtractor, dataCatalog, pub, store)
	ingestionsService = ingestions.NewLoggingMiddleware(logger, ingestionsService)

	// Initiate Endpoints
	ingestionsEndpoints := ingestions.NewEndpointSet(ingestionsService)

	// Initiate Transport Handlers
	var (
		ingestionsHTTPHandler = ingestions.NewHTTPHandler(ingestionsEndpoints, logger)
		docsHTTPHandler       = docs.NewDocsHandler()
	)

	g, gCtx := errgroup.WithContext(ctx)

	mux := chi.NewMux()
	mux.Mount("/swagger", docsHTTPHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// TODO: divide into liveness and readiness
		w.Write([]byte("ok"))
	})
	// TODO: split into /sources and /connectors
	mux.Mount("/ingestions", ingestionsHTTPHandler)
	srv := &http.Server{
		Addr:    config.HTTPServerHost,
		Handler: mux,
	}

	g.Go(func() error {
		level.Info(logger).Log("msg", "listening", "host", config.HTTPServerHost)
		return srv.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		level.Info(logger).Log("msg", "gracefully closing http server")
		return srv.Shutdown(shutDownCtx)
	})

	// r, err := message.NewRouter(message.RouterConfig{}, brokerLogger)
	// if err != nil {
	// 	panic(err)
	// }
	// connector.NewKafkaResponsesHandler(connectorEndpoints, sub, r)
	//
	// g.Go(func() error {
	// 	return r.Run(gCtx)
	// })

	err = g.Wait()
	if err != nil && err != http.ErrServerClosed {
		level.Error(logger).Log("msg", err)
	}
}
