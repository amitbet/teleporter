package main

import (
	"math/rand"
	"time"
)

//https://siongui.github.io/2015/04/13/go-generate-random-string/
func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

//writes some shiny new json
// func updatejson() {
// 	if len(clients) == 0 {
// 		clientjson = []byte("[]")
// 		return
// 	}
// 	slice := make([]map[string]string, len(clients))
// 	i := 0
// 	for _, v := range clients {
// 		slice[i] = map[string]string{"username": v.Username, "password": v.Password, "remoteip": v.Remoteip, "port": v.Port}
// 		i++
// 	}

// 	j, err := json.Marshal(slice)
// 	if err != nil {
// 		logger.Error(err)
// 		clientjson = []byte("[]")
// 		return
// 	}
// 	clientjson = j
// }

//http://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals-in-golang
func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}
