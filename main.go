package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/ulfox/kcadm/utils"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	logger *logrus.Logger
	port   string
	addr   string
)

func main() {
	cert := flag.String("cert", "./certs/server-cert.pem", "x509 certifiacte")
	key := flag.String("key", "./certs/server-key.pem", "x509 certificate key")
	listenPort := flag.String("port", "8080", "Http listening port")
	listenAddr := flag.String("addr", "0.0.0.0", "Http listening address")

	flag.Parse()

	port = *listenPort
	addr = *listenAddr

	logger = logrus.New()
	log := logger.WithFields(logrus.Fields{
		"Component": "kcadm",
	})

	log.Info("Initiating")

	certs, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:      addr + ":" + port,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}},
	}

	r := mux.NewRouter()
	r.HandleFunc("/namespace-admition", validate)
	http.Handle("/", r)
	srv.Handler = r

	sysSigs := utils.NewOSSignal()
	var httpServerWaitGroup sync.WaitGroup

	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	log.Infof("Server running listening in port: %s", port)

	sysSigs.Wait()
	log.Infof("Interrupted")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	httpServerWaitGroup.Add(1)

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Info("Bye!")

	httpServerWaitGroup.Done()
}

func httpErr(w http.ResponseWriter, m string) {
	log := logger.WithFields(logrus.Fields{
		"Component": "http-error",
	})

	log.Error(m)
	http.Error(w, m, http.StatusBadRequest)
}

func validate(w http.ResponseWriter, r *http.Request) {
	log := logger.WithFields(logrus.Fields{
		"Component": "validate-request",
	})
	log.Info(r.URL.Path)

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		httpErr(w, "Body payload is required")
		return
	}

	admissionRequest := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionRequest); err != nil {
		httpErr(w, "Invalid payload: "+err.Error())
		return
	}

	namespace := v1.Namespace{}
	if err := json.Unmarshal(admissionRequest.Request.Object.Raw, &namespace); err != nil {
		httpErr(w, err.Error())
		return
	}

	var admit bool = true
	var result metav1.Status

	if !strings.HasPrefix(namespace.Name, "user-") {
		admit = false
		result = metav1.Status{
			Code:    403,
			Message: "Namespace Name must start with user-",
		}
	}

	admissionResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: admit,
			Result:  &result,
			UID:     admissionRequest.Request.UID,
		},
	}

	admissionResponse.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"})
	encAdmissionResponse, err := json.Marshal(admissionResponse)
	if err != nil {
		httpErr(w, err.Error())
		return
	}

	if _, err := w.Write(encAdmissionResponse); err != nil {
		log.Error(err)
	}
}
