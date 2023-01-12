package cookiejar

import (
	"bytes"
	"encoding/gob"
	"time"
)

// ------------------------------------------------------------------------

// entry is the internal representation of a cookie.
// This struct type is not used outside of this package per se, but the exported
// fields are those of RFC 6265.
type entry struct {
	Name       string    `json:"name" bson:"name,omitempty"`
	Value      string    `json:"value" bson:"value,omitempty"`
	Domain     string    `json:"domain" bson:"domain,omitempty"`
	Path       string    `json:"path" bson:"path,omitempty"`
	SameSite   string    `json:"same_site" bson:"same_site,omitempty"`
	Secure     bool      `json:"secure" bson:"secure,omitempty"`
	HttpOnly   bool      `json:"http_only" bson:"http_only,omitempty"`
	Persistent bool      `json:"persistent" bson:"persistent,omitempty"`
	HostOnly   bool      `json:"host_only" bson:"host_only,omitempty"`
	Expires    time.Time `json:"expires" bson:"expires,omitempty"`
	Creation   time.Time `json:"creation" bson:"creation,omitempty"`
	LastAccess time.Time `json:"last_access" bson:"last_access,omitempty"`

	// seqNum is a sequence number so that Cookies returns cookies in a
	// deterministic order, even for cookies that have equal Path length and
	// equal Creation time. This simplifies testing.
	seqNum uint64 `json:"seq_num" bson:"seq_num,omitempty"`
}

// entries is the internal representation of a submap.
type entries map[string]entry

// ------------------------------------------------------------------------

// DecodeBinaryToEntries encodes the entry submap to bytes.
func DecodeBinaryToEntries(data []byte) (entries, error) {
	// Convert byte slice to io.Reader
	reader := bytes.NewReader(data)

	// Decode to a slice of cookies
	var e entries
	err := gob.NewDecoder(reader).Decode(&e)

	return e, err
}

// ------------------------------------------------------------------------

// BinaryEncode encodes the entry submap to bytes.
func (e entries) BinaryEncode() ([]byte, error) {
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(e)

	return b.Bytes(), err
}
