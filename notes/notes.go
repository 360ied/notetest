package notes

import (
	"bytes"
	"compress/gzip"
	"io"
	"notetest/strmap"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"

	cryptorand "crypto/rand"
)

type Notes struct {
	mu sync.RWMutex

	nm map[string]string
}

func NewEmptyDB() *Notes {
	return &Notes{
		nm: map[string]string{},
	}
}

// UnlockDB opens a Notes file
// may read more than necessary from the io.Reader
func UnlockDB(r io.Reader, key []byte) (*Notes, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(gr); err != nil {
		return nil, err
	}

	nonce := [24]byte{}
	if _, err := io.ReadFull(buf, nonce[:]); err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	pt, err := aead.Open(nil, nonce[:], buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	nm, err := strmap.Unmarshal(bytes.NewReader(pt))
	if err != nil {
		return nil, err
	}

	return &Notes{
		nm: nm,
	}, nil
}

func (n *Notes) SaveDB(w io.Writer, key []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return err
	}

	// log.Printf("AEADNONCELEN %s\n", aead.NonceSize())

	nonce := [24]byte{}
	if _, err := cryptorand.Read(nonce[:]); err != nil {
		return err
	}

	if _, err := gw.Write(nonce[:]); err != nil {
		return err
	}

	ptBuf := &bytes.Buffer{}
	strmap.Marshal(ptBuf, n.nm)

	ct := aead.Seal(nil, nonce[:], ptBuf.Bytes(), nil)

	if _, err := gw.Write(ct); err != nil {
		return err
	}

	if err := gw.Close(); err != nil {
		return err
	}

	return nil
}

type NotesUpdate struct {
	// The name of the note
	Name string `json:"name"`

	// The new content of the note
	Content string `json:"content"`

	// Whether to delete the note or not
	// If `delete` is specified, `content` is ignored
	Delete bool `json:"delete"`
}

func (n *Notes) UpdateNote(nu NotesUpdate) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if nu.Delete {
		delete(n.nm, nu.Name)
	} else {
		n.nm[nu.Name] = nu.Content
	}
}

func (n *Notes) ViewNote(name string) (string, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	content, found := n.nm[name]
	return content, found
}

func (n *Notes) ListNotes() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ret := make([]string, len(n.nm))
	for k, _ := range n.nm {
		ret = append(ret, k)
	}

	return ret
}
