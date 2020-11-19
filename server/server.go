package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx"
)

type chatServer struct {
	bind       string
	dbAddr     string
	log        io.Writer
	db         *pgx.Conn
	members    map[time.Time][]string
	connString string
}

type message struct {
	Username  string
	Password  string
	Msg       string
	timestamp time.Time
}

type user struct {
	Username string
	Password string
}

func (c *chatServer) resetConn() {
	var err error
	c.db.Close(context.Background())
	c.db, err = pgx.Connect(context.Background(), c.connString)
	if err != nil {
		fmt.Fprintln(c.log, err)
	}
}
func NewServer(logger io.Writer, bind string, connString string) chatServer {
	return chatServer{
		bind:       bind,
		log:        logger,
		connString: connString,
	}

}

func (c *chatServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "\nRooms Available: ")
	fmt.Fprintf(c.log, "\nRecieved request")
}

//func (c *chatServer)
func (c *chatServer) insertMember(m *message) {
	defer c.resetConn()
	payload := fmt.Sprintf("insert into members(username,password) values ('%v','%v')", m.Username, m.Password)
	_, err := c.db.Exec(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(c.log, "Error creating member record")
		fmt.Fprintln(c.log, err)
		os.Exit(1)
	}
	fmt.Fprintln(c.log, "created member record: ", m.Username)
}

func (c *chatServer) sendMessage(w http.ResponseWriter, r *http.Request) {
	var m message
	json.NewDecoder(r.Body).Decode(&m)
	fmt.Fprintln(w, "Request recieved from: ", m.Username)
	reqPayload := fmt.Sprintf("select username from members;")
	request, err := c.db.Query(context.Background(), reqPayload)
	if err != nil {
		fmt.Fprintln(c.log, err)
		os.Exit(1)
	}
	var o string
	for request.Next() {
		fmt.Fprintln(c.log, "scanning")
		request.Scan(&o)
		if o == m.Username {
			c.resetConn()
			m.timestamp = time.Now()
			c.insertMessage(m)
			return
		}
		c.resetConn()
	}

	c.insertMember(&m)
	//payload := fmt.Sprintf("insert into members(username,password) values ('%v','%v')", m.Username, m.Password)

}

func (c *chatServer) insertMessage(m message) (err error) {
	payload := fmt.Sprintf("insert into messages (msg,sent,username) VALUES ('%v','%v','%v')", m.Msg, m.timestamp.UTC().Format("2006-01-02T15:04:05-0700"), m.Username)
	_, err = c.db.Exec(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(c.log, err)
	}
	return
}
func (c *chatServer) Start() {
	fmt.Fprintln(c.log, "\nStarting on: ", c.bind)
	var err error
	c.db, err = pgx.Connect(context.Background(), c.connString)
	if err != nil {
		fmt.Fprintln(c.log, err)
	}
	http.HandleFunc("/", c.rootHandler)
	http.HandleFunc("/send", c.sendMessage)

	_, err = c.db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS members (username varchar(45) NOT NULL,password varchar(450) NOT NULL,enabled integer NOT NULL DEFAULT '1', PRIMARY KEY (username));")

	if err != nil {
		fmt.Fprintln(c.log, err.Error())
		os.Exit(1)
	}
	_, err = c.db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS messages (msg text NOT NULL, sent timestamp without time zone NOT NULL, username varchar(45) NOT NULL);")
	if err != nil {
		fmt.Fprintln(c.log, err.Error())
		os.Exit(1)
	}
	request, err := c.db.Query(context.Background(), "SELECT username FROM members;")
	fmt.Fprintln(c.log, "Table created", request.Err())
	var o string
	for request.Next() {
		request.Scan(&o)
		fmt.Fprintln(c.log, o)

	}
	c.resetConn()
	http.ListenAndServe(c.bind, nil)

}
