package main

import (
	"log"
	"net"
)

func main() {
	commChan := make(chan string)
	deadChan := make(chan int)
	vehiclesChan := make(chan Vehicle)
	go Route(vehiclesChan, deadChan, commChan)
	go AwaitVehicle(vehiclesChan, deadChan)

	AwaitDriver(commChan)
}

func AwaitVehicle(vehiclesChan chan Vehicle, deadChan chan int) {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Panic(err)
	}

	nextId := 0

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}

		c := make(chan string)
		vehicle := Vehicle{nextId, conn, c, deadChan}
		nextId ++

		vehiclesChan <- vehicle
		go func(v Vehicle) {
			v.Run()
		}(vehicle)
		// go Vehiclize(conn, vehicle)
	}	
}

func AwaitDriver(driverChan chan string) {
	ln, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go Drive(conn, driverChan)
	}	
}

func Route(vehiclesChan chan Vehicle, deadChan chan int, instructions chan string) {
	vehicles := make(map[int]Vehicle)

	defer func() {
        if err := recover(); err != nil {
            log.Println("failure in routing:", err)
        }
    }()

	for {
		select {
		case instruction := <- instructions:
			for _, vehicle := range vehicles {
				vehicle.commChan <- instruction
			}
		case vehicle := <-vehiclesChan:
			vehicles[vehicle.id] = vehicle
		
		case deadId := <- deadChan:
			delete(vehicles, deadId)
		}
	}
}

type Vehicle struct {
	id int
	clientConn net.Conn
	commChan chan string
	deadChan chan int
}

func (v *Vehicle) Run() {
	log.Println("New vehicle created with", v.clientConn.RemoteAddr())
	for message := range v.commChan {
		_, err := v.clientConn.Write([]byte(message))

		if err != nil {
			log.Println(err)
			v.deadChan <- v.id
			return
		}
	}
}

func Drive(driverConn net.Conn, commChan chan string) {
	log.Println("New driver connection made with", driverConn.RemoteAddr())

	for {
		data := make([]byte, 512)
		_, err := driverConn.Read(data)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Data received from driver:", string(data))
		commChan <- string(data)
	}
}

