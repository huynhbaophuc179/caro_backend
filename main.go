package main

import (
	wsData "caro_backend/data"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var hubs = make(map[string]*Hub)

func printMap(m map[string]*Hub) {
	for k, v := range m {
		fmt.Printf("%s: %v\n", k, v)
	}
}
func handleJSONResponse(w http.ResponseWriter, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// handle error
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func createServer(w http.ResponseWriter, r *http.Request) {
	fmt.Print("create server\n")
	printMap(hubs)
	hub := newHub()

	fmt.Print(wsData.HelloWord())
	generatedId := generateID()

	hubs[generatedId] = hub

	go hub.run()

	fmt.Print(generatedId)
	message := wsData.Response{
		Status:  200,
		Data:    generatedId,
		Message: "success",
	}
	// send the generated UUID to the client
	handleJSONResponse(w, message)

}
func cleanupHubs() {
	for {
		// Sleep for some interval before checking the hubs.
		time.Sleep(10 * time.Minute)

		// Iterate over all hubs and delete any that haven't been used in the last hour.
		for id, hub := range hubs {
			if time.Since(hub.lastUsed) > time.Hour {
				delete(hubs, id)
				log.Printf("Deleted hub %s", id)
			}
		}
	}
}

func main() {
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	r := mux.NewRouter()
	r.HandleFunc("/createHub", createServer).Methods("GET")
	r.HandleFunc("/ws/{id}", func(w http.ResponseWriter, r *http.Request) {
		printMap(hubs)
		vars := mux.Vars(r)
		id := vars["id"]
		hub, ok := hubs[id]

		if ok {
			fmt.Println("joined server")
			serveWs(hub, w, r)
		} else {
			createServer(w, r)
		}

	})

	// Start the hub cleanup goroutine.
	go cleanupHubs()

	port := ":8080"
	fmt.Print("Server is running on port " + port)

	err := http.ListenAndServe(port, corsHandler(r))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// generate unique id for each client using uuid
func generateID() string {
	id := uuid.New()
	return id.String()
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	lastUsed time.Time
}

// newHub create new pointer of Hub
func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		lastUsed:   time.Now(),
	}
}

// this the the machine. This function runs forever
// waiting for any new channel IO
func (h *Hub) run() {
	for {
		select {
		//handle new client connection
		case client := <-h.register:
			h.clients[client] = true

		//handle disconnected client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		//handle message broadcast
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// create a function that recive a 2 dimensional array and 2 points
// then return the shortest path between the 2 points

func checkWinner(array *[][]int, pos []int) int {
	checkDiagonalWinner := func(array *[][]int, pos []int) int {
		x := pos[0]
		y := pos[1]
		val := (*array)[x][y]
		var count int

		// check diagonal from top left to bottom right
		count = 1
		i := x - 1
		j := y - 1
		for i >= 0 && j >= 0 && (*array)[i][j] == val {
			count++
			i--
			j--
		}
		i = x + 1
		j = y + 1
		for i < len(*array) && j < len((*array)[0]) && (*array)[i][j] == val {
			count++
			i++
			j++
		}
		if count >= 5 {
			return val
		}

		// check diagonal from top right to bottom left
		count = 1
		i = x - 1
		j = y + 1
		for i >= 0 && j < len((*array)[0]) && (*array)[i][j] == val {
			count++
			i--
			j++
		}
		i = x + 1
		j = y - 1
		for i < len(*array) && j >= 0 && (*array)[i][j] == val {
			count++
			i++
			j--
		}
		if count >= 5 {
			return val
		}

		return 0
	}

	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++")
	// Struct to represent a point in 2 dimensional array

	// this function determine if there is any winner between the 2 points in the
	// 2 dimensional array or not, 0 is non, 1 is x and 2 is o
	deterWinner := func(x1 int, x2 int, y1 int, y2 int) int {
		fmt.Println("---------------------------------")
		diffX := 0
		diffY := 0
		if x1 > x2 {
			diffX = -1
		} else if x1 < x2 {
			diffX = 1
		}
		if y1 > y2 {
			diffY = -1
		} else if y1 < y2 {
			diffY = 1
		}
		xCount := 0
		oCount := 0
		iteration := 0
		//loging everything
		fmt.Println("Point 1:", x1, y1)
		fmt.Println("Point 2:", x2, y2)
		//flag for deter if the loop should continue or not
		flag := true

		for flag {
			fmt.Println("+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+_+")
			fmt.Println("iteration:", iteration)
			fmt.Println("x1, y1", x1, y1)
			fmt.Println("Point 2:", x2, y2)
			fmt.Println("maplo", (*array)[x1][y1])
			iteration++

			switch (*array)[x1][y1] {
			case 1:
				xCount++
				oCount = 0
			case 2:
				oCount++
				xCount = 0
			default:
				xCount = 0
				oCount = 0
			}
			x1 += diffX
			y1 += diffY
			if xCount == 5 {
				return 1
			}
			if oCount == 5 {
				return 2
			}
			if x1 == x2 && y1 == y2 {
				// if the loop reach the end of the line
				flag = false
				// run one more time
				switch (*array)[x1][y1] {
				case 1:
					xCount++
					oCount = 0
				case 2:
					oCount++
					xCount = 0
				default:
					xCount = 0
					oCount = 0
				}

				if xCount == 5 {
					return 1
				}
				if oCount == 5 {
					return 2
				}
			}

		}
		fmt.Print("count")
		fmt.Println(xCount, oCount)
		fmt.Println("---------------------------------")

		return 0
	}
	floorVal := func(x int) int {
		res := x
		for i := 0; i < 4; i++ {
			if res == 0 {
				break
			}
			res -= 1

		}
		return res
	}
	ceilingVal := func(x int) int {
		res := x
		for i := 0; i < 4; i++ {
			if res == 12 {
				break
			}
			res += 1

		}
		return res
	}
	x := pos[0]
	y := pos[1]

	xn := floorVal(x)
	yn := floorVal(y)
	xp := ceilingVal(x)
	yp := ceilingVal(y)
	ch := make(chan int)

	// print every points
	fmt.Println("x, y", x, y)
	fmt.Println("xn, yn", xn, yn)
	fmt.Println("xp, yp", xp, yp)

	go func() {
		ch <- deterWinner(xn, xp, y, y)
	}()

	go func() {
		ch <- deterWinner(x, x, yn, yp)
	}()

	go func() {
		ch <- checkDiagonalWinner(array, pos)
	}()

	fmt.Println("Winner deter+++++++++++++++++++++++++++++++")
	results := make([]int, 3)
	for i := 0; i < 3; i++ {
		results[i] = <-ch
	}
	close(ch)

	// Check results to determine winner
	for _, res := range results {
		if res != 0 {
			return res
		}
	}
	fmt.Println("No one win")
	return 0
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var msg wsData.Message

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error decoding json: %v", err)
			continue
		}
		fmt.Println(msg.Type)
		if msg.Type == "board" {
			fmt.Println(reflect.TypeOf(msg.Board))
			res := strings.Split(msg.Pos, ",")
			fmt.Println(msg.Pos)
			fmt.Println(res[0], "+", res[1])
			intArr := make([]int, len(res))
			for i, v := range res {
				intVal, err := strconv.Atoi(v)
				if err != nil {
					fmt.Printf("Error converting string %s to int: %v\n", v, err)
					return
				}
				intArr[i] = intVal
			}
			fmt.Println(msg.Board)
			fmt.Println(msg.Board[intArr[0]][intArr[1]])
			winner := checkWinner(&msg.Board, intArr)
			fmt.Print("winner")
			fmt.Println(winner)
			type responseData struct {
				Winner int     `json:"winner"`
				Board  [][]int `json:"board"`
			}

			resData := wsData.Response{
				Status:  200,
				Message: "OK",
				Data:    responseData{winner, msg.Board},
				Type:    "board",
			}
			byteSlice, err := json.Marshal(resData)
			if err != nil {
				log.Printf("error encoding json: %v", err)

			}
			c.hub.broadcast <- byteSlice
		} else {
			c.hub.broadcast <- message
		}
	}

}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}
	c.hub.lastUsed = time.Now()
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	hub.lastUsed = time.Now()
	go client.writePump()
	go client.readPump()
}
