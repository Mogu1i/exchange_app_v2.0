package utils

import "github.com/gorilla/websocket"

//client表示单个连接的用户
type Client struct {
	Conn     *websocket.Conn
	Username string
	Send     chan []byte
}

//hub管理所有客户端的连接
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

var Chathub = 
