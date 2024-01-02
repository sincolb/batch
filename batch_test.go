package batchrequests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func getHttpTestServer(wg *sync.WaitGroup, batch *dispatch) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &Request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			log.Println("err = ", err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*5000)
		defer cancel()
		// req.Ctx = r.Context()
		req.Ctx = ctx
		req.Value = struct {
			w http.ResponseWriter
			r *http.Request
		}{
			w: w,
			r: r,
		}
		task, err := batch.Submit(req)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		if _, err := task.Result(); err != nil {
			w.Write([]byte(err.Error()))
		}
		wg.Done()
	}))
	return ts
}

func sendRequest(ser *httptest.Server, max int) {
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			payload, _ := json.Marshal(Request{
				Id:    "key#" + strconv.Itoa(rand.Intn(max)),
				Value: rand.Intn(i + 1),
			})

			res, err := ser.Client().Post(ser.URL, "application/json; charset=UTF-8", bytes.NewReader(payload))
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				log.Fatalf("%d", res.StatusCode)
			}

			result, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("The Result = %s\n", result)
		}()
	}
}

func TestSingle(t *testing.T) {
	defer goleak.VerifyNone(t)

	batch := NewDispatch()
	defer batch.Release()

	max := 5
	wg := &sync.WaitGroup{}
	for i := 0; i < max; i++ {
		i := i
		handle := func(ctx context.Context, task *Task) bool {
			payload := task.Value.(struct {
				w http.ResponseWriter
				r *http.Request
			})
			_, err := payload.w.Write([]byte(strconv.Itoa(i) + "-surccess: " + task.Id))
			if err != nil {
				log.Println("write error: ", err)
			}
			// payload.w.Write([]byte{'\n', '\r'})
			// log.Printf("%p\n", payload.w)
			return true
		}
		batch.Register("key#"+strconv.Itoa(i), 3, time.Millisecond*3000, HandleSingle(handle))
	}

	ts := getHttpTestServer(wg, batch)
	defer ts.Close()

	wg.Add(10)
	sendRequest(ts, max)
	wg.Wait()
}

func TestBatch(t *testing.T) {
	defer goleak.VerifyNone(t)

	max := 5
	batch := NewDispatch()
	defer batch.Release()

	wg := &sync.WaitGroup{}

	for i := 0; i < max; i++ {
		i := i
		handle := func(ctx context.Context, task []*Task) bool {
			for _, item := range task {
				payload := item.Value.(struct {
					w http.ResponseWriter
					r *http.Request
				})
				_, err := payload.w.Write([]byte(strconv.Itoa(i) + "-surccess: " + item.Id))
				if err != nil {
					log.Println("write error: ", err)
				}
			}
			return true
		}
		batch.Register("key#"+strconv.Itoa(i), 3, time.Millisecond*2000, HandleBatch(handle))
	}

	ts := getHttpTestServer(wg, batch)
	defer ts.Close()

	wg.Add(10)
	sendRequest(ts, max)
	wg.Wait()
}

func BenchmarkBatch(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	batch := NewDispatch()

	index := 10
	handle := func(ctx context.Context, task []*Task) bool {
		// time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		// if rand.Intn(10)%3 == 0 {
		logger.Infof("[task] %v \n", task)
		return true
		// }
		// logger.Tracef("[%s] \n", work.uniqID())
		// return false
	}
	for i := 0; i < index; i++ {
		batch.Register("key#"+strconv.Itoa(i), 10, time.Second, HandleBatch(handle))
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		data := Request{
			Ctx:   context.Background(),
			Id:    "key#" + strconv.Itoa(rand.Intn(index)),
			Value: rand.Intn(100),
		}
		_, err := batch.Submit(&data)
		if err != nil {
			log.Println("submit err: ", err)
		}
	}
	defer batch.Release()
}
