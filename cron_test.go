package hcron

import (
	"testing"
	"context"
	"sync"
	"time"
)

type mockTask struct {
	ctx      context.Context
	i        int
	chResult chan int
}

func (m mockTask) Task(ts time.Time) {
	select {
	case <-m.ctx.Done():
	case m.chResult <- m.i:
	}
}

func TestCron(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	cron := NewCron(ctx, &wg, 1)

	chResult := make(chan int)
	ctxTask, _ := context.WithCancel(context.Background())

	times := make([]time.Time, 6)
	times[0] = time.Now().Add(time.Millisecond * 100)
	for i := 1; i < 6; i++ {
		times[i] = times[i-1].Add(1)
	}
	cron.Add(ctxTask, times[0], mockTask{ctx: ctxTask, i: 0, chResult: chResult}.Task)
	cron.Add(ctxTask, times[3], mockTask{ctx: ctxTask, i: 1, chResult: chResult}.Task)
	cron.Add(ctxTask, times[2], mockTask{ctx: ctxTask, i: 2, chResult: chResult}.Task)
	cron.Add(ctxTask, times[5], mockTask{ctx: ctxTask, i: 3, chResult: chResult}.Task)
	cron.Add(ctxTask, times[1], mockTask{ctx: ctxTask, i: 4, chResult: chResult}.Task)
	cron.Add(ctxTask, times[1], mockTask{ctx: ctxTask, i: 4, chResult: chResult}.Task)

	select {
	case i := <-chResult:
		t.Fatal("unexpected received result befor timer expired : ", i)
	case <-time.After(time.Millisecond * 10):
	}

	time.Sleep(time.Millisecond * 100)

	result := []int{0, 4, 4, 2, 1, 3}
	for _, i := range result {
		select {
		case r := <-chResult:
			if r != i {
				t.Errorf("result received but not euqal %v != %v", r, i)
			}
		case <-time.After(time.Millisecond * 10):
			t.Fatal("result waited timeout")
		}
	}
	cancel()
	wg.Wait()
}
