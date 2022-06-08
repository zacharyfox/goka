package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/lovoo/goka/examples/3-messaging/blocker"
	"github.com/lovoo/goka/examples/3-messaging/collector"
	"github.com/lovoo/goka/examples/3-messaging/detector"
	"github.com/lovoo/goka/examples/3-messaging/filter"
	"github.com/lovoo/goka/examples/3-messaging/translator"
	"golang.org/x/sync/errgroup"
)

var (
	brokers       = []string{"localhost:9092"}
	runFilter     = flag.Bool("filter", false, "run filter processor")
	runCollector  = flag.Bool("collector", false, "run collector processor")
	runTranslator = flag.Bool("translator", false, "run translator processor")
	runBlocker    = flag.Bool("blocker", false, "run blocker processor")
	runDetector   = flag.Bool("detector", false, "run detector processor")
	broker        = flag.String("broker", "localhost:9092", "boostrap Kafka broker")
)

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx := errgroup.WithContext(ctx)

	// When this example is run the first time, wait for creation of all internal topics
	initialized := sync.WaitGroup{}
	if *runCollector {
		initialized.Add(1)
	}
	if *runFilter {
		initialized.Add(1)
	}
	if *runBlocker {
		initialized.Add(1)
	}
	if *runDetector {
		initialized.Add(1)
	}
	if *runTranslator {
		initialized.Add(1)
	}

	if *runCollector {
		log.Println("starting collector")
		grp.Go(collector.Run(ctx, brokers, &initialized))
	}
	if *runFilter {
		log.Println("starting filter")
		grp.Go(filter.Run(ctx, brokers, &initialized))
	}
	if *runBlocker {
		log.Println("starting blocker")
		grp.Go(blocker.Run(ctx, brokers, &initialized))
	}
	if *runDetector {
		log.Println("starting detector")
		grp.Go(detector.Run(ctx, brokers, &initialized))
	}
	if *runTranslator {
		log.Println("starting translator")
		grp.Go(translator.Run(ctx, brokers, &initialized))
	}

	// Wait for SIGINT/SIGTERM
	waiter := make(chan os.Signal, 1)
	signal.Notify(waiter, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-waiter:
	case <-ctx.Done():
	}
	cancel()
	if err := grp.Wait(); err != nil {
		log.Println(err)
	}
	log.Println("done")
}
