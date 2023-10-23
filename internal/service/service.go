package service

import (
	"errors"
	"fmt"
	"math/rand"
	"nodasofttest/pkg"
	"nodasofttest/pkg/pool"
	"sync"
	"time"
)

func WithPools() {
	wg := &sync.WaitGroup{}
	defer pkg.Duration(pkg.Track("Duration"))

	poolCreating := pool.NewPool(5)
	poolDoing := pool.NewPool(5)
	poolSorting := pool.NewPool(5)

	resultCreating := make(chan WorkTask, 10)
	defer close(resultCreating)
	resultDoing := make(chan WorkTask, 10)
	end := make(chan bool)
	doneChan := make(chan WorkTask, 10)
	errorChan := make(chan error, 10)
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			task := pool.NewTask(creating, &ArgsCreate{resultCreating})
			poolCreating.AddTask(task)
		}
		//poolCreating.Stop()
	}(wg)

	go func(chan WorkTask) {

		for res := range resultCreating {
			go func(res WorkTask) {
				task := pool.NewTask(doing, &ArgsDoing{res, resultDoing})
				poolDoing.AddTask(task)
			}(res)
		}
		poolDoing.Stop()
	}(resultCreating)

	go func(chan WorkTask) {
		for res := range resultDoing {
			go func(res WorkTask) {
				task := pool.NewTask(sorting, &ArgsSorting{res, doneChan, errorChan})
				poolSorting.AddTask(task)
			}(res)
		}
		poolSorting.Stop()
	}(resultDoing)

	go poolCreating.Run("создание")

	go poolDoing.Run("обработка")

	go poolSorting.Run("Сортировка")

	fmt.Println("Прием результатов")
	var doneTasks []*WorkTask
	var errors []*error
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for task := range resultCreating {
			fmt.Printf("[main] Task %d has been finished with result %v \n", task.id, task.result)
			doneTasks = append(doneTasks, &task)
		}
		close(resultCreating)
	}(wg)
	go func(chan WorkTask, chan error) {

		for {
			select {
			case task := <-resultCreating:

				fmt.Printf("Done Task %d has been finished with result %v \n", task.id, task.result)
				doneTasks = append(doneTasks, &task)
			case err := <-errorChan:
				fmt.Printf("Error Task has been finished with error: %v \n", err)
				//errors = append(errors, &err)
			case <-end:
				break
			default:
				//	fmt.Printf("pizdec!!!! \n")
				//close(resultCreating)
				//return

			}
		}
	}(doneChan, errorChan)

	go func(chan WorkTask, chan error) {
		for {
			select {
			case task := <-doneChan:

				fmt.Printf("[done] Task %d has been finished with result %v \n", task.id, task.result)
				doneTasks = append(doneTasks, &task)
			case err := <-errorChan:
				fmt.Printf("[error] Task has been finished with error: %v \n", err)
				errors = append(errors, &err)
				//default:
				//	fmt.Printf("pizdec!!!! \n")
			}
		}
	}(doneChan, errorChan)
	//
	fmt.Println("Done tasks:", len(doneTasks))
	//
	fmt.Println("Undone tasks:", len(errors))
	wg.Wait()
	// deadlock из-за остановки
}

type WorkTask struct {
	id         int64
	creatTime  time.Time
	finishTime time.Time
	result     error
}

type ArgsCreate struct {
	Out chan<- WorkTask
}

func creating(args interface{}) error {
	// TODO: сделать проверки интерфейса на наличие аргументов
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	arg := args.(*ArgsCreate)

	created := time.Now()
	if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
		created = time.Time{}
	}
	id := time.Now().Unix()
	fmt.Printf("Task %d created\n", id)
	arg.Out <- WorkTask{id: id, creatTime: created}
	return nil
}

type ArgsDoing struct {
	task WorkTask
	Out  chan<- WorkTask
}

func doing(args interface{}) error {
	// TODO: сделать проверки интерфейса на наличие аргументов
	time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)

	arg := args.(*ArgsDoing)
	if !arg.task.creatTime.After(time.Now().Add(-20 * time.Second)) {
		arg.task.result = errors.New("something went wrong")
	}

	arg.task.finishTime = time.Now()

	fmt.Printf("Task %d proccesseed\n", arg.task.id)

	arg.Out <- arg.task

	return nil
}

type ArgsSorting struct {
	task     WorkTask
	outDone  chan<- WorkTask
	outError chan<- error
}

func sorting(args interface{}) error {
	// TODO: сделать проверки интерфейса на наличие аргументов
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	arg := args.(*ArgsSorting)

	fmt.Printf("Task %d sorted\n", arg.task.id)
	if arg.task.result != nil {
		err := fmt.Errorf("Task id %d time %s, error %s", arg.task.id, arg.task.creatTime, arg.task.result)
		arg.outError <- err
	} else {
		arg.outDone <- arg.task
	}

	return nil
}
