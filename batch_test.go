package batch

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

type myStruct struct {
	w http.ResponseWriter
	r *http.Request
}

func getHttpTestServer(wg *sync.WaitGroup, batch *dispatch[myStruct]) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := make(map[string]string)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("err = ", err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*5000)
		defer cancel()
		// req.Ctx = r.Context()
		value := myStruct{
			w: w,
			r: r,
		}
		task, err := batch.SubmitWithContext(ctx, req["Id"], value)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		if err := task.Wait(); err != nil {
			w.Write([]byte(err.Error()))
		}
		wg.Done()
	}))
	return ts
}

func sendRequest(ser *httptest.Server, max int) {
	for i := 0; i < 10; i++ {
		go func() {
			payload, _ := json.Marshal(map[string]string{
				"Id": "key#" + strconv.Itoa(rand.Intn(max)),
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

	batch := NewDispatch[myStruct]()
	defer batch.Release()

	max := 5
	wg := &sync.WaitGroup{}
	for i := 0; i < max; i++ {
		i := i
		handle := func(ctx context.Context, payload *myStruct) bool {
			_, err := payload.w.Write([]byte(strconv.Itoa(i)))
			if err != nil {
				log.Println("write error: ", err)
			}
			return true
		}
		batch.Register("key#"+strconv.Itoa(i), 3, time.Millisecond*3000, HandleSingle[myStruct](handle))
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
	batch := NewDispatch[myStruct]()
	defer batch.Release()

	wg := &sync.WaitGroup{}

	for i := 0; i < max; i++ {
		i := i
		handle := func(ctx context.Context, payload []*myStruct) bool {
			for _, item := range payload {
				_, err := item.w.Write([]byte(strconv.Itoa(i)))
				if err != nil {
					log.Println("write error: ", err)
				}
			}
			return true
		}
		batch.Register("key#"+strconv.Itoa(i), 3, time.Millisecond*2000, HandleBatch[myStruct](handle))
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
	batch := NewDispatch[any]()

	index := 10
	handle := func(ctx context.Context, task []*any) bool {
		// time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		// if rand.Intn(10)%3 == 0 {
		// log.Printf("[task] %v \n", len(task))
		return true
		// }
		// log.Printf("[%s] \n", work.uniqID())
		// return false
	}
	for i := 0; i < index; i++ {
		batch.Register("key#"+strconv.Itoa(i), 10, time.Second, HandleBatch[any](handle))
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := "key#" + strconv.Itoa(rand.Intn(index))
		value := rand.Intn(100)
		_, err := batch.Submit(key, value)
		if err != nil {
			log.Println("submit err: ", err)
		}
	}
	defer batch.Release()
}
