package realtime

import (
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/keanutaufan/kvstored/api/entity"
)

type SocketServer struct {
	Server  *socketio.Server
	keySubs map[string]map[string]map[string]socketio.Conn // appID -> key -> clientID -> connection
	appSubs map[string]map[string]socketio.Conn            // appID -> clientID -> connection
	mu      sync.Mutex
}

func NewSocketServer() *SocketServer {
	s := &SocketServer{
		Server:  socketio.NewServer(nil),
		keySubs: make(map[string]map[string]map[string]socketio.Conn),
		appSubs: make(map[string]map[string]socketio.Conn),
	}

	s.Server.OnConnect("/", func(so socketio.Conn) error {
		log.Println("Connected:", so.ID())
		return nil
	})

	// Modified to accept both appID and key
	s.Server.OnEvent("/", "subscribe_key", func(so socketio.Conn, appID, key string) {
		log.Printf("Client %s subscribed to key: %s in app: %s", so.ID(), key, appID)
		s.mu.Lock()
		defer s.mu.Unlock()

		// Initialize nested maps if they don't exist
		if s.keySubs[appID] == nil {
			s.keySubs[appID] = make(map[string]map[string]socketio.Conn)
		}
		if s.keySubs[appID][key] == nil {
			s.keySubs[appID][key] = make(map[string]socketio.Conn)
		}
		s.keySubs[appID][key][so.ID()] = so
	})

	s.Server.OnEvent("/", "unsubscribe_key", func(so socketio.Conn, appID, key string) {
		log.Printf("Client %s unsubscribed from key: %s in app: %s", so.ID(), key, appID)
		s.mu.Lock()
		defer s.mu.Unlock()
		if appSubs, ok := s.keySubs[appID]; ok {
			if clients, ok := appSubs[key]; ok {
				delete(clients, so.ID())
				if len(clients) == 0 {
					delete(appSubs, key)
				}
				if len(appSubs) == 0 {
					delete(s.keySubs, appID)
				}
			}
		}
	})

	s.Server.OnEvent("/", "subscribe_app", func(so socketio.Conn, appID string) {
		log.Printf("Client %s subscribed to app: %s", so.ID(), appID)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.appSubs[appID] == nil {
			s.appSubs[appID] = make(map[string]socketio.Conn)
		}
		s.appSubs[appID][so.ID()] = so
	})

	s.Server.OnEvent("/", "unsubscribe_app", func(so socketio.Conn, appID string) {
		log.Printf("Client %s unsubscribed from app: %s", so.ID(), appID)
		s.mu.Lock()
		defer s.mu.Unlock()
		if clients, ok := s.appSubs[appID]; ok {
			delete(clients, so.ID())
			if len(clients) == 0 {
				delete(s.appSubs, appID)
			}
		}
	})

	s.Server.OnDisconnect("/", func(so socketio.Conn, reason string) {
		s.mu.Lock()
		defer s.mu.Unlock()
		// Clean up key subscriptions
		for appID, appSubs := range s.keySubs {
			for key, clients := range appSubs {
				delete(clients, so.ID())
				if len(clients) == 0 {
					delete(appSubs, key)
				}
			}
			if len(appSubs) == 0 {
				delete(s.keySubs, appID)
			}
		}
		// Clean up app subscriptions
		for appID, clients := range s.appSubs {
			delete(clients, so.ID())
			if len(clients) == 0 {
				delete(s.appSubs, appID)
			}
		}
	})

	return s
}

func (s *SocketServer) NotifyKeySet(keyValue entity.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Notify key-specific subscribers
	if appSubs, ok := s.keySubs[keyValue.AppID]; ok {
		if clients, ok := appSubs[keyValue.Key]; ok {
			for _, so := range clients {
				so.Emit("key_set", keyValue)
			}
		}
	}

	// Notify app subscribers
	if clients, ok := s.appSubs[keyValue.AppID]; ok {
		for _, so := range clients {
			so.Emit("key_set", keyValue)
		}
	}
}

func (s *SocketServer) NotifyKeyUpdated(keyValue entity.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Notify key-specific subscribers
	if appSubs, ok := s.keySubs[keyValue.AppID]; ok {
		if clients, ok := appSubs[keyValue.Key]; ok {
			for _, so := range clients {
				so.Emit("key_updated", keyValue)
			}
		}
	}

	// Notify app subscribers
	if clients, ok := s.appSubs[keyValue.AppID]; ok {
		for _, so := range clients {
			so.Emit("key_updated", keyValue)
		}
	}
}

func (s *SocketServer) NotifyKeyDeleted(appID, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Notify key-specific subscribers
	if appSubs, ok := s.keySubs[appID]; ok {
		if clients, ok := appSubs[key]; ok {
			for _, so := range clients {
				so.Emit("key_deleted", gin.H{
					"app_id": appID,
					"key":    key,
				})
			}
		}
	}

	// Notify app subscribers
	if clients, ok := s.appSubs[appID]; ok {
		for _, so := range clients {
			so.Emit("key_deleted", gin.H{
				"app_id": appID,
				"key":    key,
			})
		}
	}
}
