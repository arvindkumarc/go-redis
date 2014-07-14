package main

import (
	"booking-engine/helpers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var totalTime time.Duration

func main() {

	helpers.InitRedisPool()

	//	file, err := os.Create("summa.pprof")
	//	if (err != nil) {
	//		log.Println(err)
	//	}

	//	pprof.StartCPUProfile(file)

	log.Println("Listening at 9000..")
	// Waiting for terminating (i use a sighandler like in vitess)
	//	terminate := make(chan os.Signal)
	//	signal.Notify(terminate, os.Interrupt)
	//
	//	go func() {
	//		<-terminate
	//
	//		pprof.StopCPUProfile()
	//		file.Close()
	//
	//		log.Printf("Server stopped")
	//		os.Exit(0)
	//	}()

	server := &http.Server{
		ReadTimeout: 4 * time.Second,
		Handler:     &MyHandler{},
		Addr:        ":9000",
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

type SeatRequest struct {
	SeatNames []string
}

type MyHandler struct {
}

func (this *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var start time.Time
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	if r.Method == "POST" {
		var seatRequest SeatRequest
		err = json.Unmarshal(content, &seatRequest)
		if err != nil {
			log.Println(err)
		}

		var synWait sync.WaitGroup

		start = time.Now()
		for _, seatNumber := range seatRequest.SeatNames {
			synWait.Add(1)
			go helpers.BlockSeat(seatNumber, &synWait)
		}

		synWait.Wait()

		w.WriteHeader(200)
		fmt.Fprintf(w, "ok")
	}

	fmt.Println(time.Since(start).Nanoseconds())

}
