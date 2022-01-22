package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/tarm/serial"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	comcast    *serial.Port
	comdata    chan []byte
}

type Client struct {
	socket net.Conn
	data   chan []byte
}

func (manager *ClientManager) start() {
	manager.startCom()
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			fmt.Println("Added new connection!")
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.data)
				delete(manager.clients, connection)
				fmt.Println("A connection has terminated!")
			}
		case message := <-manager.broadcast:
			for connection := range manager.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(manager.clients, connection)
				}
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
			msg := message[0:length]
			//manager.broadcast <- msg
			manager.comdata <- msg
		}
	}
}

func (manager *ClientManager) startCom() {
	// read from the data source
	config := &serial.Config{Name: "COM ~ Port", Baud: 19200, ReadTimeout: 0}
	s, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	manager.comcast = s
}

func (manager *ClientManager) rcvcom() {
	for {
		message := make([]byte, 4096)
		length, err := manager.comcast.Read(message)
		if err != nil {
			manager.comcast.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED COM: " + string(message))
			manager.broadcast <- message[0:length]
		}
	}
}

func (manager *ClientManager) sendcom() {
	defer manager.comcast.Close()
	for {
		select {
		case message, ok := <-manager.comdata:
			if !ok {
				return
			}
			manager.comcast.Write(message)
		}
	}
}

func (client *Client) receive() {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			client.socket.Write(message)
		}
	}
}

func startServerMode() {
	fmt.Println("Starting server...")
	listener, error := net.Listen("tcp", ":12345")
	if error != nil {
		fmt.Println(error)
	}
	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		comdata:    make(chan []byte),
	}
	go manager.start()
	for {
		connection, _ := listener.Accept()
		if error != nil {
			fmt.Println(error)
		}

		// 최대 접속자 제한.
		if len(manager.clients) > 0 {
			connection.Close()
			fmt.Println("이미 접속되어 있습니다.")
			continue
		}

		client := &Client{socket: connection, data: make(chan []byte, 10)} // 자주 데이터를 받는 경우 채널 허용량을 늘려야 한다.
		manager.register <- client
		go manager.receive(client)
		go manager.send(client)
		go manager.rcvcom()
		go manager.sendcom()
	}
}

// 클라이언트 테스트.
func startClientMode() {
	fmt.Println("Starting client...")
	connection, error := net.Dial("tcp", "localhost:12345")
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{socket: connection}
	go client.receive()
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		connection.Write([]byte(strings.TrimRight(message, "\n")))
	}
}

// end
func main() {
	startServerMode()
}
