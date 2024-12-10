package caddy_test

import (
	"fmt"
	"github.com/caddyserver/caddy/v2/caddytest"
	"net/http"
	"path/filepath"
	"testing"
)

func TestRestartWorkerViaAdminApi(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(`
		{
			skip_install_trust
			admin localhost:2999
			http_port `+testPort+`

			frankenphp {
				worker ../testdata/worker-with-counter.php 1
			}
		}

		localhost:`+testPort+` {
			route {
				root ../testdata
				rewrite worker-with-counter.php
				php
			}
		}
		`, "caddyfile")

	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:1")
	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:2")

	assertAdminResponse(tester, "POST", "workers/restart", http.StatusOK, "workers restarted successfully\n")

	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:1")
}

func TestRemoveWorkerThreadsViaAdminApi(t *testing.T) {
	absWorkerPath, _ := filepath.Abs("../testdata/worker-with-counter.php")
	tester := caddytest.NewTester(t)
	tester.InitServer(`
		{
			skip_install_trust
			admin localhost:2999
			http_port `+testPort+`

			frankenphp {
				worker ../testdata/worker-with-counter.php 4
			}
		}

		localhost:`+testPort+` {
			route {
				root ../testdata
				rewrite worker-with-counter.php
				php
			}
		}
		`, "caddyfile")

	// make a request to the worker to make sure it's running
	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:1")

	// remove a thread
	expectedMessage := fmt.Sprintf("New thread count: 3 %s\n", absWorkerPath)
	assertAdminResponse(tester, "POST", "workers/remove", http.StatusOK, expectedMessage)

	// remove 2 threads
	expectedMessage = fmt.Sprintf("New thread count: 1 %s\n", absWorkerPath)
	assertAdminResponse(tester, "POST", "workers/remove?count=2", http.StatusOK, expectedMessage)

	// get 400 status if removing the last thread
	assertAdminResponse(tester, "POST", "workers/remove", http.StatusBadRequest, "")

	// make a request to the worker to make sure it's still running
	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:2")
}

func TestAddWorkerThreadsViaAdminApi(t *testing.T) {
	absWorkerPath, _ := filepath.Abs("../testdata/worker-with-counter.php")
	tester := caddytest.NewTester(t)
	tester.InitServer(`
		{
			skip_install_trust
			admin localhost:2999
			http_port `+testPort+`

			frankenphp {
				worker ../testdata/worker-with-counter.php 1
			}
		}

		localhost:`+testPort+` {
			route {
				root ../testdata
				rewrite worker-with-counter.php
				php
			}
		}
		`, "caddyfile")

	// make a request to the worker to make sure it's running
	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:1")

	// get 400 status if the filename is wrong
	assertAdminResponse(tester, "POST", "workers/add?file=wrong.php", http.StatusBadRequest, "")

	// add a thread
	expectedMessage := fmt.Sprintf("New thread count: 2 %s\n", absWorkerPath)
	assertAdminResponse(tester, "POST", "workers/add", http.StatusOK, expectedMessage)

	// add 2 threads
	expectedMessage = fmt.Sprintf("New thread count: 4 %s\n", absWorkerPath)
	assertAdminResponse(tester, "POST", "workers/add?count=2", http.StatusOK, expectedMessage)

	// get 400 status if adding too many threads
	assertAdminResponse(tester, "POST", "workers/add?count=100", http.StatusBadRequest, "")

	// make a request to the worker to make sure it's still running
	tester.AssertGetResponse("http://localhost:"+testPort+"/", http.StatusOK, "requests:2")
}

func TestShowTheCorrectThreadDebugStatus(t *testing.T) {
	absWorker1Path, _ := filepath.Abs("../testdata/worker-with-counter.php")
	absWorker2Path, _ := filepath.Abs("../testdata/index.php")
	tester := caddytest.NewTester(t)
	tester.InitServer(`
		{
			skip_install_trust
			admin localhost:2999
			http_port `+testPort+`

			frankenphp {
				num_threads 6
				max_threads 12
				worker ../testdata/worker-with-counter.php 2
				worker ../testdata/index.php 2
			}
		}

		localhost:`+testPort+` {
			route {
				root ../testdata
				rewrite worker-with-counter.php
				php
			}
		}
		`, "caddyfile")

	assertAdminResponse(tester, "POST", "workers/remove?file=index.php", http.StatusOK, "")
	assertAdminResponse(tester, "POST", "threads/remove", http.StatusOK, "")

	// assert that all threads are in the right state via debug message
	assertAdminResponse(
		tester,
		"GET",
		"threads/status",
		http.StatusOK, `Thread 0 (ready) Regular PHP Thread
Thread 1 (inactive)
Thread 2 (ready) Worker PHP Thread - `+absWorker1Path+`
Thread 3 (ready) Worker PHP Thread - `+absWorker1Path+`
Thread 4 (ready) Worker PHP Thread - `+absWorker2Path+`
Thread 5 (inactive)
6 additional threads can be started at runtime
`,
	)
}

func assertAdminResponse(tester *caddytest.Tester, method string, path string, expectedStatus int, expectedBody string) {
	adminUrl := "http://localhost:2999/frankenphp/"
	r, err := http.NewRequest(method, adminUrl+path, nil)
	if err != nil {
		panic(err)
	}
	if expectedBody == "" {
		tester.AssertResponseCode(r, expectedStatus)
	} else {
		tester.AssertResponse(r, expectedStatus, expectedBody)
	}
}