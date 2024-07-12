/*
===================================================================================================================
1. Стадия фильтрации отрицательных чисел (не пропускать отрицательные числа).
2. Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0.
3.Стадия буферизации данных в кольцевом буфере с интерфейсом, соответствующим тому, который был дан в качестве задания в 19 модуле.
В этой стадии предусмотреть опустошение буфера (и соответственно, передачу этих данных, если они есть, дальше)
с определённым интервалом во времени. Значения размера буфера и этого интервала времени сделать настраиваемыми
(как мы делали: через константы или глобальные переменные).
4. Написать источник данных для конвейера. Непосредственным источником данных должна быть консоль.
5.Также написать код потребителя данных конвейера. Данные от конвейера можно направить снова в консоль построчно,
сопроводив их каким-нибудь поясняющим текстом, например: «Получены данные …».
6.При написании источника данных подумайте о фильтрации нечисловых данных, которые можно ввести через консоль.
Как и где их фильтровать, решайте сами.
===================================================================================================================
*/

package main

import (
	"container/ring"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	BuferSize    = 10
	BuferTimeOut = time.Second
)

// ***STAGE 1***
func StageFilterNegative(done <-chan struct{}, in <-chan int) <-chan int {
	//fmt.Println("StageFilterNegative started")
	out := make(chan int)

	go func() {
		defer close(out)
		for v := range in {
			if v > 0 {
				log.Printf("Enetered number is positive: %v\n", v)
				select {
				case out <- v:
				case <-done:
					log.Println("StageFilterNegative DONE")
					return
				}
			}
		}
	}()
	return out
}

// ***STAGE 2***
func StageFilterThree(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			if v%3 == 0 && v != 0 {
				log.Printf("Entered number is multiple of 3: %v\n", v)
				select {
				case out <- v:
				case <-done:
					log.Println("StageFilterThree DONE")
					return
				}
			}
		}

	}()
	return out
}

// ***STAGE 3***
func StageBufer(done <-chan struct{}, r *ring.Ring, in ...<-chan int) {
	var wg sync.WaitGroup
	out := make(chan interface{})

	Bufer := func(c <-chan int) {
		defer close(out)
		defer wg.Done()
		for v := range c {
			r.Value = v
			out <- r.Value
			r = r.Next()
		}
	}
	wg.Add(len(in))
	go func() {
		for _, v := range in {
			go Bufer(v)
		}
	}()

	for v := range out {
		time.Sleep(BuferTimeOut)
		r = r.Next()
		log.Printf("Data from the conveyor has been received: %v\n", v)
	}
	go func() {
		wg.Wait()
	}()

}

func main() {
	r := ring.New(BuferSize)
	done := make(chan struct{})
	defer close(done)

	var integers int

	init := func(integers ...int) <-chan int {
		output := make(chan int)
		go func() {
			for _, i := range integers {
				defer close(output)
				output <- i
			}
		}()
		return output
	}
	// TODO:filter chars here. ONLY NUMBERS!
	log.Printf("PIPELINE STARTED || Bufer size: %v || Bufer timeout: %v\n", BuferSize, BuferTimeOut)
	fmt.Println("Enter data for pipeline: ")
	for {
		_, err := fmt.Scan(&integers)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Entered number: %v\n", integers)
		StageBufer(done, r, StageFilterThree(done, StageFilterNegative(done, init(integers))))
	}
}
