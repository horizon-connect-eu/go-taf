package completionhandler

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCompletionHandler(t *testing.T) {
	ch := New(func() {
		fmt.Println("ok")
	}, func(err error) {
		fmt.Println("fail: " + err.Error())
	})

	var wg sync.WaitGroup

	wg.Add(1)
	resolve, _ := ch.Register()
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		resolve()
	}()

	wg.Add(1)
	resolve, _ = ch.Register()
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		resolve()
	}()

	wg.Add(1)
	_, reject := ch.Register()
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		reject(errors.New("Error!"))
	}()

	wg.Wait()
	time.Sleep(2 * time.Second)
	ch.Execute()

}
