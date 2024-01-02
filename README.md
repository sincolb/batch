Batch requests and timeout automatically commit.

```
batch := NewDispatch()
defer batch.Release()

index := 10
handle := func(ctx context.Context, task []*Task) bool {
    fmt.Printf("[task] %v \n", task)
    return true
}
for i := 0; i < index; i++ {
    batch.Register("key#"+strconv.Itoa(i), 10, time.Second, HandleBatch(handle))
}
for i := 0; i < index; i++ {
    data := Request{
        Ctx:   context.Background(),
        Id:    "key#" + strconv.Itoa(rand.Intn(index)),
        Value: rand.Int(),
    }
    _, err := batch.Submit(&data)
    if err != nil {
        fmt.Println("submit err: ", err)
    }
}
```
OR
```
batch := NewDispatch()
defer batch.Release()

index := 10
handle := func(ctx context.Context, task *Task) bool {
    fmt.Printf("[task] %v \n", task)
    return true
}
for i := 0; i < index; i++ {
    batch.Register("key#"+strconv.Itoa(i), 10, time.Second, HandleSingle(handle))
}
for i := 0; i < index; i++ {
    data := Request{
        Ctx:   context.Background(),
        Id:    "key#" + strconv.Itoa(rand.Intn(index)),
        Value: rand.Int(),
    }
    _, err := batch.Submit(&data)
    if err != nil {
        fmt.Println("submit err: ", err)
    }
}
```
