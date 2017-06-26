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
	"sync"
	"syscall"
)

var (
	opts   *Options
	logger *log.Logger
)

type protocol interface {
	run()
	shutdown()
}

func main() {
	var (
		wg       sync.WaitGroup
		signalCh = make(chan os.Signal, 1)
	)

	opts = GetOptions()

	opts.Logger.Println("In here 1") // TODO_REMOVE
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	sFlow := NewSFlow()

	opts.Logger.Println("In here 2") // TODO_REMOVE
	ipfix := NewIPFIX()

	opts.Logger.Println("In here 3") // TODO_REMOVE
	netflow9 := NewNetflowV9()

	opts.Logger.Println("In here 4") // TODO_REMOVE

	protocols := []protocol{sFlow, ipfix, netflow9}

	opts.Logger.Println("In here 5") // TODO_REMOVE
	for _, p := range protocols {
		wg.Add(1)
		go func(p protocol) {
			defer wg.Done()
			p.run()
		}(p)
	}

	opts.Logger.Println("In here 6") // TODO_REMOVE
	go statsHTTPServer(ipfix, sFlow, netflow9)

	opts.Logger.Println("In here 7") // TODO_REMOVE
	<-signalCh

	opts.Logger.Println("In here 8") // TODO_REMOVE
	for _, p := range protocols {
		wg.Add(1)
		go func(p protocol) {
			defer wg.Done()
			p.shutdown()
		}(p)
	}

	wg.Wait()
}
