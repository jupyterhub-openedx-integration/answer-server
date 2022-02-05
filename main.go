// Copyright <2021> (see CONTRIBUTERS file)
// for license, see LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/abbot/go-http-auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const version = "1.0.1"

var (
	secret   = flag.String("secret", "", "Default secret for use during testing")
	consumer = flag.String("consumer", "", "Default consumer")
	// host        = flag.String("host", "", "host name")
	// port        = flat.String("port", "", "port number")
	httpAddress = flag.String("https", "www.mathtech.org:5001", "Listen to")
)

func main() {
	flag.Parse()
	log.Printf("Lis %s, waiting request. v%s", *httpAddress, version)

	// init the database
	err := initGlobalStore("answers.db")
	if err != nil {
		log.Fatal(Err(err, "couldn't not acquire database"))
	}

	// these next two files are not included in the repository. At the
	// moment, they are manually generated by running a python script
	// against the acme.json file used by Traefik which was generated
	// by the let's encrypt registration process.

	keyFile := "./key.pem"
	certFile := "./certificate.pem"

	authenticator := auth.NewBasicAuthenticator(*httpAddress, secretPass)

	http.HandleFunc("/get-answers", authenticator.Wrap(getJupDataHandler))
	http.HandleFunc("/submit-answers", authenticator.Wrap(submitAnswerHandler))
	log.Fatal(http.ListenAndServeTLS(*httpAddress, certFile, keyFile, nil))
}

// This function is passed up to the authenticator library for
// authentication purposes.
func secretPass(user, realm string) string {
	if user == *consumer {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*secret), bcrypt.DefaultCost)
		if err == nil {
			return string(hashedPassword)
		}
	}
	return ""
}

func submitAnswerHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	log.Println("got request for submit-answers")
	// the submission server in the kubernetes cluster will be
	// originating the requests in this handler.
	edxAnonId := r.FormValue("edx-anon-id")
	// TODO introduce "labname" across magic, xblock, submit server.
	labName := r.FormValue("labname")
	labAnswerBytes := r.FormValue("lab-answers")

	globalStore.InsertAnswer(edxAnonId, labName, string(labAnswerBytes))
	log.Println(edxAnonId, labName, string(labAnswerBytes))
}

func getJupDataHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	// Grab the use userid and answerpath from the request.

	// userid will look like: jupyter-35dd7e9124c8847ec5-030ef
	edxAnonId := r.FormValue("edx-anon-id")

	// example answerpath: jupyter-answer-magic/test/test_autograde1.json
	labName := r.FormValue("labname")

	if edxAnonId == "" {
		fmt.Fprintf(w, `{"error": "edAnonId missing from form data"}`)
	}
	if labName == "" {
		fmt.Fprintf(w, `{"error": "labname missing from form data"}`)
	}

	sub, err := globalStore.GetAnswers(edxAnonId, labName)
	if err != nil {
		log.Println(err)

		switch err {
		case gorm.ErrRecordNotFound:
			msg := fmt.Sprintf("answers could not be found for this lab: %s, %s", labName, err.Error())
			fmt.Fprintf(w, `{"error": %s"}`, msg)
		default:
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
		}
	} else {
		fmt.Fprintf(w, sub.LabAnswers)
	}
}
