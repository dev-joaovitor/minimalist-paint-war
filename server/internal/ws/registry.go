package ws

import "sync"

// MemRegistry is a minimal Registrar used until the hub takes over (milestone 3).
// It enforces one active connection per username and echoes pings.
type MemRegistry struct {
	mu      sync.Mutex
	clients map[string]*Client // username -> client
}

// NewMemRegistry creates an empty registry.
func NewMemRegistry() *MemRegistry {
	return &MemRegistry{clients: make(map[string]*Client)}
}

// Register admits a client unless its username is already connected.
func (m *MemRegistry) Register(c *Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.clients[c.Username]; exists {
		return ErrUsernameTaken
	}
	m.clients[c.Username] = c
	c.SendMsg(TypeJoined, JoinedData{
		UserID:   c.ID,
		Username: c.Username,
		Role:     "LOBBY_PLAYER",
		Team:     "",
	})
	return nil
}

// Unregister removes a client if it is the one currently registered.
func (m *MemRegistry) Unregister(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur, ok := m.clients[c.Username]; ok && cur == c {
		delete(m.clients, c.Username)
	}
}

// HandleMessage replies to pings; other messages are ignored for now.
func (m *MemRegistry) HandleMessage(c *Client, env Envelope) {
	if env.Type == TypePing {
		c.SendMsg(TypePong, env.Data)
	}
}
