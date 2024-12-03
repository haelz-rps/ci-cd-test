package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"

	"github.com/Schub-cloud/security-bulletins/api/feed"
	feedpb "github.com/Schub-cloud/security-bulletins/api/pb/feed"
	"github.com/Schub-cloud/security-bulletins/api/repository"
	"github.com/Schub-cloud/security-bulletins/api/repository/gormimpl"
	"github.com/Schub-cloud/security-bulletins/api/utils"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kafka "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	httpServerHostDefault    = ""
	brokerHostDefault        = "localhost"
	httpServerPortDefault    = "8080"
	grpcServerPortDefault    = "8081"
	metricsServerPortDefault = "9091"
	logLevelDefault          = "debug"
	requestsTopicDefault     = "api-requests"
	kafkaGroupIDDefault      = "api"
)

func main() {
	ctx := context.Background()

	// Load server configs
	var (
		dbConnString      = utils.BuildDBConnectionString()
		serverHost        = utils.EnvString("HOST", httpServerHostDefault)
		brokerHost        = utils.EnvString("BROKER_HOST", brokerHostDefault)
		httpServerPort    = utils.EnvString("PORT", httpServerPortDefault)
		grpcServerPort    = utils.EnvString("GRPC_PORT", grpcServerPortDefault)
		metricsServerPort = utils.EnvString("METRICS_PORT", metricsServerPortDefault)
		requestsTopic     = utils.EnvString("REQUESTS_TOPIC", requestsTopicDefault)
		kafkaGroupID      = utils.EnvString("KAFKA_GROUP_ID", kafkaGroupIDDefault)

		httpServerFullHost    = fmt.Sprintf("%s:%s", serverHost, httpServerPort)
		grpcServerFullHost    = fmt.Sprintf("%s:%s", serverHost, grpcServerPort)
		metricsServerFullHost = fmt.Sprintf("%s:%s", serverHost, metricsServerPort)

		logLevel       = strings.ToLower(utils.EnvString("LOG_LEVEL", logLevelDefault))
		parsedLogLevel = level.Allow(level.ParseDefault(logLevel, level.InfoValue()))
	)

	// Init logger
	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = level.NewFilter(logger, parsedLogLevel)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	httpLogger := log.With(logger, "component", "http")
	grpcLogger := log.With(logger, "component", "grpc")
	httpSrvLogger := log.With(logger, "transport", "http", "address", httpServerFullHost)
	grpcSrvLogger := log.With(logger, "transport", "grpc", "address", grpcServerFullHost)
	kafkaLogger := log.With(logger, "transport", "kafka")
	metricsSrvLogger := log.With(logger, "transport", "metrics-http", "address", metricsServerFullHost)

	// Init repositories
	db := repository.NewGormRepo()
	if err := db.Dial(ctx, dbConnString); err != nil {
		level.Error(logger).Log(err)
	}

	var (
		feedRepo = gormimpl.NewFeedRepository(db.DB)
		postRepo = gormimpl.NewPostsRepository(db.DB)
	)

	promFieldKeys := []string{"method"}

	// Init Services
	var feedSvc feed.FeedService
	{
		feedSvc = feed.NewFeedService(feedRepo, postRepo)
		feedSvc = feed.NewLoggingMiddleware(logger, feedSvc)
		feedSvc = feed.NewMetricsMiddleware(
			kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "api",
				Subsystem: "feed_service",
				Name:      "request_count",
				Help:      "Number of requests per method.",
			}, promFieldKeys),
			kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
				Namespace: "api",
				Subsystem: "feed_service",
				Name:      "request_latency_microseconds",
				Help:      "Total duration of requests in microseconds.",
			}, promFieldKeys),
			feedSvc,
		)
	}
	// Init Endpoints
	var (
		feedEndpointSet = feed.NewEndpointSet(feedSvc)
	)

	// Init Transport Handlers
	var (
		metricsHandler   = promhttp.Handler()
		feedHTTPHandler  = feed.NewHTTPHandler(feedEndpointSet, httpLogger)
		feedGRPCServer   = feed.NewGRPCServer(*feedEndpointSet, grpcLogger)
		feedKafkaHandler = feed.NewKafkaMessageHandler(feedEndpointSet, kafkaLogger)
	)

	var g run.Group
	{
		// HTTP Server
		mux := http.NewServeMux()
		mux.Handle("/feeds", feedHTTPHandler)
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		httpListener, err := net.Listen("tcp", httpServerFullHost)
		if err != nil {
			level.Error(httpLogger).Log("during", "Listen", "err", err)
			os.Exit(1)
		}

		g.Add(func() error {
			level.Info(httpSrvLogger).Log("msg", "listening")
			return http.Serve(httpListener, mux)
		}, func(error) {
			level.Info(httpSrvLogger).Log("msg", "closing")
			err := httpListener.Close()
			if err != nil {
				level.Error(httpSrvLogger).Log("msg", err)
			} else {
				level.Info(httpSrvLogger).Log("msg", "closed")
			}
		})
	}
	{
		// Kafka
		kafkaReader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{brokerHost},
			GroupID:     kafkaGroupID,
			GroupTopics: []string{requestsTopic},
		})
		kafkaWriter := kafka.Writer{
			Addr: kafka.TCP(brokerHost),
		}

		level.Info(kafkaLogger).Log("msg", "start listening kafka messages", "topics", fmt.Sprintf("%s", requestsTopic))
		g.Add(func() error {
			for {
				r := kafka.Message{}
				m, err := kafkaReader.ReadMessage(ctx)
				if err != nil {
					level.Error(kafkaLogger).Log(err)
					return err
				}
				feedKafkaHandler(ctx, m, &r)
				if r.Topic != "" {
					err = kafkaWriter.WriteMessages(ctx, r)
					if err != nil {
						return err
					}
				}
			}
		}, func(err error) {
			level.Info(kafkaLogger).Log("msg", "closing")
			kafkaReader.Close()
			kafkaWriter.Close()
			level.Info(kafkaLogger).Log("msg", "cosed")
		})
	}
	{
		// gRPC Server
		srvMetrics := grpcprom.NewServerMetrics(
			grpcprom.WithServerHandlingTimeHistogram(
				grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
			),
		)
		reg := stdprometheus.NewRegistry()
		reg.MustRegister(srvMetrics)

		baseServer := grpc.NewServer(
			grpc.UnaryInterceptor(kitgrpc.Interceptor),
			grpc.ChainUnaryInterceptor(
				srvMetrics.UnaryServerInterceptor(),
			),
		)
		g.Add(func() error {
			grpcListener, err := net.Listen("tcp", grpcServerFullHost)
			if err != nil {
				level.Error(grpcSrvLogger).Log("during", "Listen", "err", err)
				os.Exit(1)
			}
			level.Info(grpcSrvLogger).Log("msg", "listening")

			srvMetrics.InitializeMetrics(baseServer)
			feedpb.RegisterFeedServiceServer(baseServer, feedGRPCServer)
			reflection.Register(baseServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			level.Info(grpcSrvLogger).Log("msg", "closing")
			baseServer.GracefulStop()
			baseServer.Stop()
			level.Info(grpcSrvLogger).Log("msg", "closed")
		})
	}
	{
		mux := http.NewServeMux()
		mux.Handle("/metrics", metricsHandler)

		httpListener, err := net.Listen("tcp", metricsServerFullHost)
		if err != nil {
			level.Error(metricsSrvLogger).Log("during", "Listen", "err", err)
			os.Exit(1)
		}

		g.Add(func() error {
			level.Info(metricsSrvLogger).Log("msg", "listening")
			return http.Serve(httpListener, mux)
		}, func(error) {
			level.Info(metricsSrvLogger).Log("msg", "closing")
			err := httpListener.Close()
			if err != nil {
				level.Error(metricsSrvLogger).Log("msg", err)
			} else {
				level.Info(metricsSrvLogger).Log("msg", "closed")
			}
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	level.Info(logger).Log("exit", g.Run())
}
