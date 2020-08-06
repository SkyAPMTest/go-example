package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SkyAPM/go2sky"
	go2skyHttp "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
)

func main() {
	svc := flag.String("svc", "", "Service name")
	ins := flag.String("ins", "", "Instance name")
	nextSvc := flag.String("next", "", "The next service")
	port := flag.Int("port", 3000, "Port to listen")
	oap := flag.String("oap", "oap:11800", "Addr of OAP server")
	flag.Parse()

	// Use gRPC reporter for production
	r, err := reporter.NewGRPCReporter(*oap)
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	tracer, err := go2sky.NewTracer(*svc, go2sky.WithReporter(r), go2sky.WithInstance(*ins))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	sm, err := go2skyHttp.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create server middleware error %v \n", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", sm(endFunc(*nextSvc, tracer)))

	log.Printf("Listening on :%d...", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
	log.Fatal(err)

}

func endFunc(nextSvc string, tracer *go2sky.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if nextSvc == "" {
			log.Printf("request method: %s\n", r.Method)
			time.Sleep(50 * time.Millisecond)
			return
		}
		client, err := go2skyHttp.NewClient(tracer)
		if err != nil {
			log.Fatalf("create client error %v \n", err)
		}


		// call next service
		request, err := http.NewRequest("POST", fmt.Sprintf("http://%s/", nextSvc), nil)
		if err != nil {
			log.Fatalf("unable to create http request: %+v\n", err)
		}
		res, err := client.Do(request)
		if err != nil {
			log.Fatalf("unable to do http request: %+v\n", err)
		}
		_ = res.Body.Close()
		
	}
}
