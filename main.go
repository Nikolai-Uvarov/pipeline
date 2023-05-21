package main

import (
	"container/ring"
	"fmt"
	"log"

	//"os"
	"strconv"
	"time"
)

// Напишите код, реализующий пайплайн, работающий с целыми числами и состоящий из следующих стадий:

// Стадия фильтрации отрицательных чисел (не пропускать отрицательные числа).
// Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0.
// Стадия буферизации данных в кольцевом буфере с интерфейсом, соответствующим тому,
// который был дан в качестве задания в 19 модуле. В этой стадии предусмотреть опустошение
// буфера (и соответственно, передачу этих данных, если они есть, дальше) с определённым
// интервалом во времени. Значения размера буфера и этого интервала времени сделать
// настраиваемыми (как мы делали: через константы или глобальные переменные).

// Написать источник данных для конвейера. Непосредственным источником данных должна быть консоль.

// Также написать код потребителя данных конвейера. Данные от конвейера можно направить снова в
// консоль построчно, сопроводив их каким-нибудь поясняющим текстом, например:
// «Получены данные …».

// При написании источника данных подумайте о фильтрации нечисловых данных, которые можно
// ввести через консоль. Как и где их фильтровать, решайте сами.

const bufSize = 5
const bufWaitSec = 10

// источник данных для пайплайна из консоли. порождает канал данных и управляющий канал остановки пайплайна
func DataSupply() (data chan int, stop chan int) {
	data = make(chan int)
	stop = make(chan int)

	go func() {
		defer close(data)

		var inp string

		//fmt.Println("Введите целые числа, после каждого \"enter\".\n Для остановки обработки введите \"stop\"")
		log.Println("Введите целые числа, после каждого \"enter\".\n Для остановки обработки введите \"stop\"")
		for {
			_, err := fmt.Scanln(&inp)
			if err != nil {
				log.Fatal("Unexpected input error:", err)
				//os.Exit(0)
			}

			if inp == "stop" {
				close(stop)
				log.Println("Pipeline stopped by user")
				break
			}

			i, err := strconv.Atoi(inp)
			if err != nil {
				log.Println("Ошибка ввода: введите число или строку \"stop\"\n можете продолжить ввод")
				continue
			}

			data <- i
		}

	}()
	return
}

//стадия пайплайна - фильтрация отрицательных чисел

func FilterNegative(data <-chan int, stop <-chan int) <-chan int {
	output := make(chan int)

	go func() {
		defer close(output)
	loop:
		for {
			select {
			case i := <-data:
				if i >= 0 {
					output <- i
					log.Println("positive: ", i, "passed filtration")
				} else {
					log.Println("filtered negative: ", i)
				}
			case <-stop:
				break loop
			}
		}
	}()

	return output
}

// Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0.
func FilterNotDivBy3(data <-chan int, stop <-chan int) <-chan int {
	output := make(chan int)

	go func() {
		defer close(output)
	loop:
		for {
			select {
			case i := <-data:
				if i != 0 && i%3 == 0 {
					output <- i
					log.Println("dividalbe by 3: ", i, "passed filtration")
				} else {
					log.Println("filtered zero or not dividalbe by 3: ", i)
				}
			case <-stop:
				break loop
			}
		}
	}()

	return output
}

// Стадия буферизации данных в кольцевом буфере с таймаутом, после которого выдача потребителю

func DataBuffer(data <-chan int, stop <-chan int) <-chan int {

	output := make(chan int)
	buf := ring.New(bufSize)
	begin := buf
	//канал-сигнал о наличии свободного места в буфере
	cap := make(chan int, bufSize)

	//принимаем данные в буфер, если есть свободная емкость
	go func() {
		defer close(cap)
	loop:
		for {
			select {
			case i := <-data:
				cap <- 1
				buf.Value = i
				buf = buf.Next()
				log.Println("Added to buffer: ", i)
			case <-stop:
				break loop
			}
		}
	}()

	//освобождаем весь буфер (в т.ч. все, что попадет в него во время самого процесса освобождения)
	go func() {
		defer close(output)
	loop:
		for {
			select {
			case <-time.Tick(time.Second * bufWaitSec):

				//опустошаем буфер
				for begin != buf || begin.Value != nil {
					i := begin.Value.(int)
					output <- i
					begin.Value = nil
					begin = begin.Next()
					<-cap
					log.Println("Send from buffer: ", i)
				}
			case <-stop:
				break loop
			}
		}
	}()

	return output

}

// потребитель данных, выводит полученые данные в консоль после получения уведомления об остановке пайплайна
// предназначена для работы в основном потоке после запуска пайплайна, чем гарантирует незавершение горутин пайплайна
// из-за завершения основного потока
func DataConsumer(data <-chan int, stop <-chan int) {

	output := make([]int, 0, 10)
	for {
		select {
		case i := <-data:
			output = append(output, i)
			log.Println("recieved by consumer: ", i)
		case <-stop:
			log.Println("Получены данные:\n", output, "\n", "Goodbye!")
			return
		}
	}
}

func main() {
	data, stop := DataSupply()

	DataConsumer(DataBuffer(FilterNotDivBy3(FilterNegative(data, stop), stop), stop), stop)
}
