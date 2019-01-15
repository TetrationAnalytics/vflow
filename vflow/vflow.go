//: ----------------------------------------------------------------------------
//: Copyright (C) 2017 Verizon.  All Rights Reserved.
//: All Rights Reserved
//:
//: file:    vflow.go
//: details: TODO
//: author:  Mehrdad Arshad Rad
//: date:    02/01/2017
//:
//: Licensed under the Apache License, Version 2.0 (the "License");
//: you may not use this file except in compliance with the License.
//: You may obtain a copy of the License at
//:
//:     http://www.apache.org/licenses/LICENSE-2.0
//:
//: Unless required by applicable law or agreed to in writing, software
//: distributed under the License is distributed on an "AS IS" BASIS,
//: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//: See the License for the specific language governing permissions and
//: limitations under the License.
//: ----------------------------------------------------------------------------

// Package main is the vflow binary
package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

var (
	opts   *Options
	logger *log.Logger
)

type proto interface {
	run()
	shutdown()
}

func main() {
	var (
		wg       sync.WaitGroup
		signalCh = make(chan os.Signal, 1)
	)

	opts = GetOptions()
	runtime.GOMAXPROCS(opts.GetCPU())

	opts.Logger.Println("In here 1") // TODO_REMOVE
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	sFlow := NewSFlow()

	opts.Logger.Println("In here 2") // TODO_REMOVE
	ipfix := NewIPFIX()
	netflow5 := NewNetflowV5()
	netflow9 := NewNetflowV9()

	protos := []proto{sFlow, ipfix, netflow5, netflow9}

	opts.Logger.Println("In here 5") // TODO_REMOVE
	for _, p := range protos {
		wg.Add(1)
		go func(p proto) {
			defer wg.Done()
			p.run()
		}(p)
	}

	go statsHTTPServer(ipfix, sFlow, netflow5, netflow9)

	opts.Logger.Println("In here 7") // TODO_REMOVE
	<-signalCh

	opts.Logger.Println("In here 8") // TODO_REMOVE
	for _, p := range proto {
		wg.Add(1)
		go func(p proto) {
			defer wg.Done()
			p.shutdown()
		}(p)
	}

	wg.Wait()
}
