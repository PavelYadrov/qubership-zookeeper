// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	zkSessionTimeout = 10 * time.Second
	certFilePath     = "/opt/zookeeper/tls/tls.crt"
	keyFilePath      = "/opt/zookeeper/tls/tls.key"
	caCertFilePath   = "/opt/zookeeper/tls/ca.crt"
)

var (
	command     = flag.String("c", "", "The command to run ZooKeeper assistant application (backup - to start backup server,  health - to perform health check)")
	source      = flag.String("s", "", "The repository where ZooKeeper stores transaction logs")
	destination = flag.String("d", "", "The repository where transaction logs should be moved")
	credentials = flag.String("u", "", "The credentials of admin ZooKeeper user")

	format = logging.MustStringFormatter("[%{time:2006-01-02T15:04:05.999}][%{level}] %{message}")
	log    = logging.MustGetLogger("main")

	verbose          = GetEnv("DEBUG", "false")
	sslEnabled       = getBoolEnv("ENABLE_SSL", "false")
	twoWaySslEnabled = getBoolEnv("ENABLE_2WAY_SSL", "false")
	g                errgroup.Group
)

type State struct {
	Status string
}

type Data struct {
	Source      string
	Destination string
}

type Result struct {
	Status  string
	Message string
}

func zookeeperConnect(servers []string, sessionTimeout time.Duration) (*zk.Conn, error) {
	if sslEnabled {
		return zookeeperConnectTLS(servers, sessionTimeout)
	}
	zkconn, _, err := zk.Connect(servers, sessionTimeout)
	return zkconn, err
}

func zookeeperConnectTLS(servers []string, sessionTimeout time.Duration) (*zk.Conn, error) {
	caCert, err := os.ReadFile(caCertFilePath)
	if err != nil {
		return nil, err
	}

	var certFile string
	var keyFile string
	if twoWaySslEnabled {
		certFile = certFilePath
		keyFile = keyFilePath
	}

	dialer, err := newTLSDialer(servers[0], caCert, certFile, keyFile)
	if err != nil {
		return nil, err
	}
	zkconn, _, err := zk.Connect(servers, sessionTimeout, zk.WithDialer(dialer))
	return zkconn, err
}

func newTLSDialer(addr string, caCert []byte, certFile, keyFile string) (zk.Dialer, error) {
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to add root certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	if len(certFile) > 0 && len(keyFile) > 0 {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, errors.New("cannot read TLS certificate or key file: " + err.Error())
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return func(string, string, time.Duration) (net.Conn, error) {
		return tls.Dial("tcp", addr, tlsConfig)
	}, nil
}

func main() {
	flag.Parse()

	// Define logging
	loggingLevel := logging.INFO
	if verbose == "true" {
		loggingLevel = logging.DEBUG
	}

	loggingSetting := logging.NewLogBackend(os.Stdout, "", 0)
	loggingLeveled := logging.AddModuleLevel(loggingSetting)
	loggingLeveled.SetLevel(loggingLevel, "")
	loggingFormatter := logging.NewBackendFormatter(loggingSetting, format)
	logging.SetBackend(loggingFormatter)

	if *command == "health" {
		c, err := zookeeperConnect([]string{"127.0.0.1:2181"}, zkSessionTimeout)
		if err != nil {
			log.Fatal("Failed to connect to ZooKeeper: ", err)
		}
		defer c.Close()

		if credentials != nil {
			err = c.AddAuth("digest", []byte(*credentials))
			if err != nil {
				log.Fatal("Failed to add authentication for ZooKeeper client: ", err)
			}
		}
		if _, _, _, err = c.ChildrenW("/"); err != nil {
			log.Fatal("Failed to obtain list of znodes: ", err)
		}
		log.Info("ZooKeeper health check is successful")
	} else if *command == "backup" {
		data := Data{
			Source:      *source,
			Destination: *destination,
		}

		server := &http.Server{
			Addr:    ":8081",
			Handler: BackupHandlers(data),
		}

		g.Go(func() error {
			return server.ListenAndServe()
		})

		if err := g.Wait(); err != nil {
			log.Fatal(err)
		}
	}
}

func BackupHandlers(data Data) http.Handler {
	r := mux.NewRouter()
	r.Handle("/", http.HandlerFunc(GetState())).Methods("GET")
	r.Handle("/store", http.HandlerFunc(data.Store())).Methods("POST")
	return JsonContentType(handlers.CompressHandler(r))
}

func GetState() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				response := State{Status: "Error"}
				w.WriteHeader(500)
				resbody, marshalerr := json.Marshal(response)
				if marshalerr != nil {
					log.Error("Failed to marshal response to json: ", marshalerr)
					return
				}
				w.Write(resbody)
				return
			}
		}()

		response := State{Status: "Running"}
		w.WriteHeader(200)
		responseBody, _ := json.Marshal(response)
		w.Write(responseBody)
	}
}

func (data Data) Store() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				response := Result{Status: "Error"}
				w.WriteHeader(500)
				resbody, marshalerr := json.Marshal(response)
				if marshalerr != nil {
					log.Error("Failed to marshal response to json: ", marshalerr)
					return
				}
				w.Write(resbody)
				return
			}
		}()

		response := storeZooKeeperFiles(data)

		w.WriteHeader(200)
		responseBody, _ := json.Marshal(response)
		log.Debugf("Response body: %s\n", responseBody)
		w.Write(responseBody)
	}
}

func storeZooKeeperFiles(data Data) Result {
	// Remove destination directory if it exists
	_, err := os.Stat(data.Destination)
	if err == nil {
		err = os.RemoveAll(data.Destination)
		if err != nil {
			return Result{Status: "Error", Message: fmt.Sprintf(
				"Remove of '%s' folder is failed: %s", data.Destination, err.Error())}
		}
	}
	// Create destination directory
	err = os.Mkdir(data.Destination, 0755)
	if err != nil {
		return Result{Status: "Error", Message: fmt.Sprintf(
			"Creation of '%s' folder is failed: %s", data.Destination, err.Error())}
	}
	// Copy files (without directory) from source to destination directory
	source := fmt.Sprintf("%s/.", data.Source)
	copyCommand := exec.Command("cp", "-rp", source, data.Destination)
	err = copyCommand.Run()
	if err != nil {
		return Result{Status: "Error",
			Message: fmt.Sprintf("Copying command is failed: %s", err.Error())}
	} else {
		return Result{Status: "Ok", Message: "The files are copied successfully!"}
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getBoolEnv(key string, defaultValue string) bool {
	return strings.ToLower(GetEnv(key, defaultValue)) == "true"
}

func JsonContentType(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}
