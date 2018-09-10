package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type Device struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Index               int    `json:"index"`
	DeviceString        string `json:"device.string"`
	DeviceIntendedRoles string `json:"device.intended_roles"`
	DeviceDesc          string `json:"device.description"`
	DeviceModel         string `json:"device.model"`
	DeviceIconName      string `json:"device.icon_name"`
}

type ZoneListItem struct {
	Name string   `json:"name"`
	Sink []Device `json:"sink"`
}

var upgrader = websocket.Upgrader{
	Subprotocols:      []string{"grut"},
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		http.Error(w, reason.Error(), status)
	},
}

var buffer = 4
var bufferLength = 44100 * buffer

func audioWS(w http.ResponseWriter, r *http.Request) {

	zone := strings.TrimPrefix(r.URL.Path, "/audio/")
	zone = strings.TrimSuffix(zone, "/socket/")
	fmt.Println("Socket for zone " + zone + " connected")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	err = c.WriteMessage(websocket.TextMessage, []byte("start"))
	if err != nil {
		log.Println("write:", err)
		return
	}
	counter := 0
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		counter += len(message)
		if counter >= bufferLength {
			err = c.WriteMessage(websocket.TextMessage, []byte("stop"))
			if err != nil {
				log.Println("write:", err)
				break
			}
			time.Sleep(5 * time.Second)
			counter = 0
			// log.Printf("recv: %s", message)
			err = c.WriteMessage(websocket.TextMessage, []byte("start"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

var d1 = Device{"Speaker1", "The first speaker", 0, "speaker_one", "music", "Sonos 1", "Sonos", "sp1.png"}
var d2 = Device{"Speaker2", "The second speaker", 1, "speaker_two", "music", "Sonos 2", "Sonos", "sp2.png"}
var d3 = Device{"Speaker3", "The third speaker", 2, "speaker_three", "music", "Sonos 3", "Sonos", "sp3.png"}
var d4 = Device{"Speaker4", "The fourth speaker", 3, "speaker_fourth", "music", "Sonos 4", "Sonos", "sp4.png"}

//GetDevice returns devices
func GetDevice(w http.ResponseWriter, r *http.Request) {
	var dl []Device
	dl = append(dl, d1)
	dl = append(dl, d2)
	dl = append(dl, d3)
	dl = append(dl, d4)
	json.NewEncoder(w).Encode(dl)
}

//GetZone
func GetZone(w http.ResponseWriter, r *http.Request) {
	var zl []ZoneListItem

	var dl1 []Device
	dl1 = append(dl1, d1)
	dl1 = append(dl1, d2)

	var dl2 []Device
	dl2 = append(dl2, d3)
	dl2 = append(dl2, d4)

	z1 := ZoneListItem{"all", dl1}
	z2 := ZoneListItem{"z1", dl2}
	zl = append(zl, z1)
	zl = append(zl, z2)
	json.NewEncoder(w).Encode(zl)
}

//SetBuffer
func SetBuffer(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	b := strings.TrimPrefix(r.URL.Path, "/audio/buffer/capacity/")
	fmt.Println("Setting buffer to " + b)
	bi, _ := strconv.Atoi(b)
	bufferLength = 44100 * bi
	w.Write([]byte("Buffer capacity set"))
}

func getRequests(w http.ResponseWriter, r *http.Request) {
	p1 := strings.TrimPrefix(r.URL.Path, "/audio/")
	switch p1 {
	case "list/device":
		fmt.Println("List device function")
		GetDevice(w, r)
		return
	case "list/zone":
		fmt.Println("List Zone function")
		GetZone(w, r)
		return
	default:
		if strings.Contains(p1, "socket") {
			fmt.Println("Websocket function")
			audioWS(w, r)
			return
		} else if strings.Contains(p1, "buffer") {
			fmt.Println("Buffer function")
			SetBuffer(w, r)
			return
		}

		return
	}
}

func main() {
	flag.Parse()
	//router := mux.NewRouter()
	http.HandleFunc("/audio/", getRequests)
	// http.HandleFunc("/audio/{zone}/socket", audioWS)
	// http.HandleFunc("/audio/list/device", getRequests)       //.Methods("GET")
	// http.HandleFunc("/audio/list/zone", getRequests)         //.Methods("GET")
	// http.HandleFunc("/audio/buffer/capacity/{b}", SetBuffer) //.Methods("POST")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
