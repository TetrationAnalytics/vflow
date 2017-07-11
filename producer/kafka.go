//: ----------------------------------------------------------------------------
//: Copyright (C) 2017 Verizon.  All Rights Reserved.
//: All Rights Reserved
//:
//: file:    kafka.go
//: details: vflow kafka producer plugin
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

package producer

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"crypto/x509"

	"github.com/Shopify/sarama"
	"gopkg.in/yaml.v2"
)

// Kafka represents kafka producer
type Kafka struct {
	producer sarama.AsyncProducer
	config   KafkaConfig
	logger   *log.Logger
}

// KafkaConfig represents kafka configuration
type KafkaConfig struct {
	Brokers      []string `yaml:"brokers"`
	Compression  string   `yaml:"compression"`
	RetryMax     int      `yaml:"retry-max"`
	RetryBackoff int      `yaml:"retry-backoff"`
	// TLS config
	SecureKafkaEnable     bool   `yaml:"secure-kafka-enable"`
	ClientCertificateFile string `yaml:"client-cert"`
	ClientPrivateKeyFile  string `yaml:"client-key"`
	RootCAFile            string `yaml:"root-ca"`
}

func (k *Kafka) setup(configFile string, logger *log.Logger) error {
	var (
		config = sarama.NewConfig()
		err    error
	)

	// set default values
	k.config = KafkaConfig{
		Brokers: []string{"localhost:9092"},
	}

	// load configuration if available
	if err = k.load(configFile); err != nil {
		logger.Println(err)
	}

	// init kafka configuration
	config.ClientID = "vFlow.Kafka"
	config.Producer.Retry.Max = k.config.RetryMax
	config.Producer.Retry.Backoff = time.Duration(k.config.RetryBackoff) * time.Millisecond

	// if secure kafka is availabe set the TLS config76y
	if k.config.SecureKafkaEnable {
		config.Net.TLS.Enable = true

		config.Net.TLS.Config = &tls.Config{}
		// set TLS version to TLSv1.2
		config.Net.TLS.Config.MinVersion = 0x0303
		if k.config.ClientCertificateFile != "" && k.config.ClientPrivateKeyFile != "" {
			logger.Println("Cert file: ", k.config.ClientCertificateFile, " Cert Key: ", k.config.ClientPrivateKeyFile)
			cert, err := tls.LoadX509KeyPair(k.config.ClientCertificateFile, k.config.ClientPrivateKeyFile)
			if err != nil {
				logger.Println("Error loading certificats: ", err)
				return err
			}
			config.Net.TLS.Config.Certificates = []tls.Certificate{cert}
		}
		if k.config.RootCAFile != "" {
			logger.Println("Root CA: ", k.config.RootCAFile)
			tlsCertPool := x509.NewCertPool()
			caCertFile, err := ioutil.ReadFile(k.config.RootCAFile)
			if err != nil {
				logger.Println("ERROR: failed to read custom Certificate Authority file", err)
				return err
			}
			if !tlsCertPool.AppendCertsFromPEM(caCertFile) {
				logger.Fatal("ERROR: failed to append certificates from Certificate Authority file")
				return nil
			}
			config.Net.TLS.Config.RootCAs = tlsCertPool
		}
	}
	switch k.config.Compression {
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	default:
		config.Producer.Compression = sarama.CompressionNone
	}

	// get env config
	if err = k.loadEnv(config); err != nil {
		logger.Println(err)
	}

	if err = config.Validate(); err != nil {
		logger.Fatal(err)
	}

	k.producer, err = sarama.NewAsyncProducer(k.config.Brokers, config)
	k.logger = logger
	if err != nil {
		return err
	}

	return nil
}

func (k *Kafka) inputMsg(topic string, mCh chan []byte, ec *uint64) {
	var (
		msg []byte
		ok  bool
	)

	k.logger.Printf("start producer: Kafka, brokers: %+v, topic: %s\n",
		k.config.Brokers, topic)

	for {
		msg, ok = <-mCh
		if !ok {
			break
		}

		select {
		case k.producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(append([]byte{}, msg...)),
		}:
		case err := <-k.producer.Errors():
			k.logger.Println(err)
			*ec++
		}
	}

	k.producer.Close()
}

func (k *Kafka) load(f string) error {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, &k.config)
	if err != nil {
		return err
	}

	return nil
}

func (k *Kafka) loadEnv(config *sarama.Config) error {
	var err error

	env := "VFLOW_KAFKA_BROKERS"
	val, ok := os.LookupEnv(env)
	if ok {
		k.config.Brokers = strings.Split(val, ";")
	}

	env = "VFLOW_KAFKA_COMPRESSION"
	val, ok = os.LookupEnv(env)
	if ok {
		k.config.Compression = val
	}

	env = "VFLOW_KAFKA_RETRY_MAX"
	val, ok = os.LookupEnv(env)
	if ok {
		k.config.RetryMax, err = strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("%s: %s", env, err)
		}
	}

	env = "VFLOW_KAFKA_RETRY_BACKOFF"
	val, ok = os.LookupEnv(env)
	if ok {
		k.config.RetryBackoff, err = strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("%s: %s", env, err)
		}
	}

	return nil
}
