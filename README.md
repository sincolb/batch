Batch requests and timeout automatically commit.

```
type myStruct struct {
    A int
    B string
}
process := batch.NewDispatch[myStruct]()
defer process.Release()

index := 10
handle := func(ctx context.Context, payload []*myStruct) bool {
    fmt.Printf("[payload] %v\n", payload)
    return true
}
for i := 0; i < index; i++ {
    process.Register("key#"+strconv.Itoa(i), 10, time.Second, batch.HandleBatch[myStruct](handle))
}
var wg sync.WaitGroup
wg.Add(index)
for i := 0; i < index; i++ {
    go func(i int) {
        defer wg.Done()

        key := "key#" + strconv.Itoa(rand.Intn(index))
        value := myStruct{A: rand.Int(), B: strconv.Itoa(i)}
        task, err := process.Submit(key, value)
        if err != nil {
            fmt.Println("submit err: ", err)
            return
        }
        err = task.Wait()
        if err != nil {
            fmt.Println(err)
            return
        }
    }(i)
}
wg.Wait()
```
OR
```
type myStruct struct {
    A int
    B string
}
process := batch.NewDispatch[myStruct]()
defer process.Release()

index := 10
handle := func(ctx context.Context, payload *myStruct) bool {
    fmt.Printf("[payload] %d %s\n", payload.A, payload.B)
    return true
}
for i := 0; i < index; i++ {
    process.Register("key#"+strconv.Itoa(i), 10, time.Second, batch.HandleSingle[myStruct](handle))
}
var wg sync.WaitGroup
wg.Add(index)
for i := 0; i < index; i++ {
    go func(i int) {
        defer wg.Done()

        key := "key#" + strconv.Itoa(rand.Intn(index))
        value := myStruct{A: rand.Int(), B: strconv.Itoa(i)}
        task, err := process.Submit(key, value)
        if err != nil {
            fmt.Println("submit err: ", err)
            return
        }
        err = task.Wait()
        if err != nil {
            fmt.Println(err)
            return
        }
    }(i)
}
wg.Wait()
```
