Batch requests and timeout automatically commit.

```
type myStruct struct {
    A int
    B string
}
batch := batchrequests.NewDispatch[myStruct]()
defer batch.Release()

index := 10
handle := func(ctx context.Context, payload []*myStruct) bool {
    fmt.Printf("[payload] %v\n", payload)
    return true
}
for i := 0; i < index; i++ {
    batch.Register("key#"+strconv.Itoa(i), 10, time.Second, batchrequests.HandleBatch[myStruct](handle))
}
var wg sync.WaitGroup
wg.Add(index)
for i := 0; i < index; i++ {
    data := batchrequests.Request[myStruct]{
        Ctx:   context.Background(),
        Id:    "key#" + strconv.Itoa(rand.Intn(index)),
        Value: myStruct{A: rand.Int(), B: strconv.Itoa(i)},
    }
    go func() {
        defer wg.Done()
        task, err := batch.Submit(&data)
        if err != nil {
            fmt.Println("submit err: ", err)
            return
        }
        err = task.Wait()
        if err != nil {
            fmt.Println(err)
            return
        }
    }()
}
wg.Wait()
```
OR
```
type myStruct struct {
    A int
    B string
}
batch := batchrequests.NewDispatch[myStruct]()
defer batch.Release()

index := 10
handle := func(ctx context.Context, payload *myStruct) bool {
    fmt.Printf("[payload] %d %s\n", payload.A, payload.B)
    return true
}
for i := 0; i < index; i++ {
    batch.Register("key#"+strconv.Itoa(i), 10, time.Second, batchrequests.HandleSingle[myStruct](handle))
}
var wg sync.WaitGroup
wg.Add(index)
for i := 0; i < index; i++ {
    data := batchrequests.Request[myStruct]{
        Ctx:   context.Background(),
        Id:    "key#" + strconv.Itoa(rand.Intn(index)),
        Value: myStruct{A: rand.Int(), B: strconv.Itoa(i)},
    }
    go func() {
        defer wg.Done()
        task, err := batch.Submit(&data)
        if err != nil {
            fmt.Println("submit err: ", err)
            return
        }
        err = task.Wait()
        if err != nil {
            fmt.Println(err)
            return
        }
    }()
}
wg.Wait()
```
