package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type TestCase struct {
	Route            string
	ExpectReturnCode int
	Conditions       map[string][]Condition
	Actions          map[string]Action
}

type TestCaseResult struct {
	Case             TestCase
	ActualReturnCode int
	ResponseTime     time.Duration
	Success          bool
	ResponseBody     string
	ErrMessages      []string
}

type TestResult struct {
	Title             string
	TestCases         []TestCaseResult
	SuccesPercentage  int
	Totaltestduration time.Duration
}

func (T *TestCaseResult) AddErrMsg(msg string) {
	T.ErrMessages = append(T.ErrMessages, msg)
}

func main() {

	testFile, err := os.Open("./testinput.json")

	if err != nil {
		panic(err)
	}

	defer testFile.Close()

	cases := []TestCase{}

	err = json.NewDecoder(testFile).Decode(&cases)

	if err != nil {
		panic(err)
	}

	caseLength := len(cases)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		result := TestResult{
			Title:     fmt.Sprintf("Testresult from %s", time.Now().Format(time.Kitchen)),
			TestCases: []TestCaseResult{},
		}

		totalTestStartTime := time.Now()

		clnt := http.Client{}

		headerAdditions := map[string]string{}
		for itNum, x := range cases {
			now := time.Now()

			testCaseResult := TestCaseResult{
				Case: x,
			}

			log.Printf("Running test %d of %d", itNum+1, caseLength)

			url := fmt.Sprintf("http://localhost:8181%s", x.Route)
			req, err := http.NewRequest(http.MethodGet, url, nil)

			if err != nil {
				testCaseResult.AddErrMsg(fmt.Sprintf("Error constructing request: %s", err.Error()))
				continue
			}

			for key, val := range headerAdditions {
				req.Header.Add(key, val)
			}

			resp, err := clnt.Do(req)
			if err != nil {
				testCaseResult.AddErrMsg(fmt.Sprintf("error sending request: %s", err.Error()))
				result.TestCases = append(result.TestCases, testCaseResult)
				continue
			}

			timeToRespond := time.Since(now)

			testCaseResult.ResponseTime = timeToRespond

			if x.ExpectReturnCode != resp.StatusCode {
				testCaseResult.AddErrMsg(fmt.Sprintf("Expected statuscode %d, got %d instead", x.ExpectReturnCode, resp.StatusCode))
			}

			respBody := make(map[string]string)

			if x.Conditions != nil {
				err = json.NewDecoder(resp.Body).Decode(&respBody)
				if err != nil {
					testCaseResult.AddErrMsg("Expected body to run conditions on, but found EOF")
				} else {
					testCaseResult.ValidateConditions(respBody)
				}

			}

			if x.Actions != nil && len(x.Actions) > 0 {
				for key, _ := range x.Actions {
					headerAdditions[key] = respBody[key]
				}
			}

			if len(testCaseResult.ErrMessages) == 0 {
				testCaseResult.Success = true
			}

			if len(respBody) > 0 {
				respBodyAsString, err := json.Marshal(respBody)
				if err == nil {
					testCaseResult.ResponseBody = string(respBodyAsString)
				}
			}

			result.TestCases = append(result.TestCases, testCaseResult)

		}

		result.Totaltestduration = time.Since(totalTestStartTime)

		// jsonOutput, _ := json.Marshal(result)

		// log.Println(string(jsonOutput))

		tmpl, err := template.ParseFiles("templates/testresult.html")

		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(w, result)
		if err != nil {
			panic(err)
		}
	})

	log.Println("Listening on port 8080")
	http.ListenAndServe("localhost:8080", nil)

}
