package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
	bank_impl "github.com/sgerogia/sol-stablecoin/tpp-client/bank/impl"
	"github.com/sgerogia/sol-stablecoin/tpp-client/cmd/config"
	contract2 "github.com/sgerogia/sol-stablecoin/tpp-client/contract"
	"github.com/sgerogia/sol-stablecoin/tpp-client/encrypt"
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	event_impl "github.com/sgerogia/sol-stablecoin/tpp-client/event/impl"
	"github.com/sgerogia/sol-stablecoin/tpp-client/schedule"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if os.Getenv("PRIVATE_KEY") == "" {
		panic(errors.New("PRIVATE_KEY env var not set"))
	}
	privateKey := os.Getenv("PRIVATE_KEY")

	// Command line flags for the service
	var configPath = flag.String("config", "./tpp-client.toml", "Location of the config TOML file.\nDefaults to ./tpp-client.toml")
	flag.Parse()

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "m",

			LevelKey:    "l",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "t",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "c",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	l, err := cfg.Build()
	if err != nil {
		panic("Unable to create logger")
	}
	logger := l.Sugar()

	// Create channel to notify the main goroutine when to stop the server.
	errc := make(chan error)

	conf, err := config.LoadConfig(*configPath, logger)
	if err != nil {
		panic("Unable to load config: " + err.Error())
	}

	// start the clients
	_, _, err = startClients(conf, privateKey, logger)
	if err != nil {
		panic("Unable to start clients: " + err.Error())
	}

	logger.Info("Listening for on-chain events...")

	// Setup interrupt handler. SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Wait for signal.
	logger.Info(fmt.Sprintf("Exiting (%v)", <-errc))

	logger.Info("Done.")
}

func startClients(
	conf *config.Config,
	privateKey string,
	l *zap.SugaredLogger,
) (*event.EventSubscriber, *gocron.Scheduler, error) {

	// get Ethereum chainClient and key pair
	chainClient, err := contract2.NewContractClient(
		conf.Ethereum.ProviderUrl,
		conf.Ethereum.ChainId,
		conf.Ethereum.ContractAddress,
		privateKey,
		conf.Ethereum.MaxGas,
	)
	if err != nil {
		return nil, nil, errors.New("Unable to create Ethereum chainClient: " + err.Error())
	}
	keyPair, err := encrypt.NewKeyPairFromHex(privateKey)
	if err != nil {
		return nil, nil, errors.New("Unable to create key pair: " + err.Error())
	}

	// get bank client
	cr := bank.OauthClientCreds{
		ClientId:       conf.BankClient.ClientId,
		ClientSecret:   conf.BankClient.ClientSecret,
		RedirectionUrl: conf.BankClient.RedirectUrl,
	}
	bankClient := bank_impl.NewNatwestSandboxClient(
		conf.Tuning.BankClientTimeout,
		&cr,
		l,
	)

	// scheduling & event handling
	sch := schedule.NewPaymentScheduler(l)
	rcv := bank.AccountDetails{
		AccountNumber: conf.BankAccount.AccountNumber,
		SortCode:      conf.BankAccount.SortCode,
		Name:          conf.BankAccount.AccountName,
	}
	handler := event_impl.NewEventHandler(
		chainClient,
		keyPair,
		bankClient,
		&rcv,
		sch,
		l)

	subscriber := event_impl.NewEventSubscriber(
		handler,
		chainClient,
		l)

	// // start the chain event subscriber
	// l.Infow("Starting chain event subscriber",
	// 	"chainId", conf.Ethereum.ChainId,
	// 	"contract", conf.Ethereum.ContractAddress)
	// ok, err := subscriber.SubscribeToMintRequestEvent()
	// if err != nil || !ok {
	// 	return nil, nil, errors.New("Unable to subscribe to MintRequestEvent: " + err.Error())
	// }

	// ok, err = subscriber.SubscribeToAuthGrantedEvent()
	// if err != nil || !ok {
	// 	return nil, nil, errors.New("Unable to subscribe to AuthGrantedEvent: " + err.Error())
	// }

	// schedule bank polling
	l.Info("Starting bank polling scheduler")
	paymentTask := schedule.NewPaymentStatusTask(sch, bankClient, handler, l)
	s := gocron.NewScheduler(time.UTC)
	s.Every(conf.Tuning.BankCronSchedule).Seconds().Do(paymentTask.CheckPaymentStatuses)

	// schedule chain polling
	l.Info("Starting contract polling scheduler")
	contractTask := schedule.NewContractEventTask(conf.Tuning.StartingBlock, chainClient, &handler, l)
	s.Every(conf.Tuning.ChainCronSchedule).Seconds().Do(contractTask.FetchAndProcessEvents)
	
	s.StartAsync()
	
	return &subscriber, s, nil
}

type chainInfo struct {
	providerUrl     string
	chainId         int64
	contractAddress string
	privateKey      string
	gasLimit        int
}
