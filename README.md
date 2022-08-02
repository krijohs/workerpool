# Workerpool

Workerpool is a library used to manage a pool of goroutines that handles incoming jobs and returns the results  
on a channel. The library supports generics so one can set it up to so the result is of the desired type.

## Usage

A new pool is initialized using the exported `New` function, which should be type instantiated with the type  
you want the result to be of. The parameter context is used for cancellation and the available options parameters  
is described below.

- `SetWorkers(workers int)` - set the number of workers, defaults to logical CPUs
- `SetJobsBuffer(size int)` - set the buffer size of the jobs channel used to store incoming jobs
- `DisableResults()` - disables writing the result to the outgoing channel returned by `Results` method

If one doesn't care about using the the returned results, `DisableResults()` option needs to be passed in to avoid results buffer being filled up, resulting in pool not being able to handle more jobs.

Example (error checks is omitted for brevity):

```go
pool := New[int]()
pool.Start(context.Background())

go func() {
	for _, i := range []int{1, 2, 3} {
		index := i
		pool.Add(context.Background(), func() (int, error) { return index, nil })
	}
	pool.Wait(context.Background())
}()

resultCh, _ := pool.Results()
for result := range resultCh {
	if result.Err != nil {
		fmt.Println(result.Err) // if an error occured while processing the job
		continue
	}
	fmt.Println(result.Result) // an int in this case
}
```

## Why

Often when needing to handle work concurrently with goroutines I tend to write similar boilerplate  
every time. Instead of rewriting the logic everytime I decided to create a module for it.

There are a lot of other workerpool libraries for Go, but I wanted to create a workerpool that uses generics  
to return the result of a processed job.
