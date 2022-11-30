package api

import (
	"crypto/md5"
	"strconv"
	"time"
)

// ValidationError represents a field that is too long
type ValidationError string

func (v ValidationError) Error() string {
	return "Field " + string(v) + " is too long"
}

func checkLength(s string, l int) bool {
	return len(s) <= l
}

// Hash is an MD5 hash
type Hash [md5.Size]byte

// NewHash returns a hashed MD5 Sum of the given strings
func NewHash(args ...string) Hash {
	var data []byte
	for _, s := range args {
		data = append(data, []byte(s)...)
	}
	return md5.Sum(data)
}

// Entry represents information about a computer
type Entry struct {
	UID              uint32    `json:"uid,omitempty"`
	Username         string    `json:"username"`
	FullName         string    `json:"full_name"`
	Serial           string    `json:"serial"`
	ClientIdentifier string    `json:"client_identifier,omitempty"`
	Hostname         string    `json:"hostname,omitempty"`
	IP               string    `json:"ip"`
	InternetIP       string    `json:"internet_ip,omitempty"`
	Time             time.Time `json:"time,omitempty"`
}

// Hashes generates and returns the user, device, address, and identity hashes for an Entry
func (e *Entry) Hashes() (user Hash, device Hash, address Hash, identity Hash) {
	u := NewHash(strconv.Itoa(int(e.UID)), e.Username, e.FullName)
	d := NewHash(e.Serial, e.ClientIdentifier, e.Hostname)
	a := NewHash(e.IP, e.InternetIP)
	i := NewHash(string(u[:]), string(d[:]), string(a[:]))
	return u, d, a, i
}

// Validate checks that the given Entry's fields fit in the DB
func (e *Entry) Validate() error {
	switch {
	case !checkLength(e.Username, 64):
		return ValidationError(e.Username)
	case !checkLength(e.FullName, 128):
		return ValidationError(e.FullName)
	case !checkLength(e.Serial, 32):
		return ValidationError(e.Serial)
	case !checkLength(e.ClientIdentifier, 64):
		return ValidationError(e.ClientIdentifier)
	case !checkLength(e.Hostname, 32):
		return ValidationError(e.Hostname)
	case !checkLength(e.IP, 15):
		return ValidationError(e.IP)
	case !checkLength(e.InternetIP, 15):
		return ValidationError(e.InternetIP)
	}
	return nil
}

// User represents a user
type User struct {
	ID       int
	UID      uint32
	Username string
	FullName string
}

// Device represents a device
type Device struct {
	ID               int
	Serial           string
	ClientIdentifier string
	Hostname         string
}

// Address represents a set of IP Addresses
type Address struct {
	ID         int
	IP         string
	InternetIP string
}

// Identity represents a unique set of User, Device, and Address
type Identity struct {
	ID        int
	UserID    int
	DeviceID  int
	AddressID int
}

// LogEntry represents an event
type LogEntry struct {
	ID         int64
	IdentityID int
	Time       time.Time
}
