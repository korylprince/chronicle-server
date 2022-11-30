package api

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

type counter int64

func (c *counter) Next() int64 {
	(*c)++
	return int64(*c)
}

func logQueryError(stmt *sql.Stmt, args ...interface{}) {
	if stmt == nil {
		log.Println("nil sql.Stmt passed to logQueryError with args:", args)
	}
	query := reflect.ValueOf(*stmt).FieldByName("query").String()
	log.Println("Error: Offending query:", fmt.Sprintf(strings.Replace(query, "?", "\"%v\"", -1), args...))
}

// Insert represents a Decomposed Entry. If a field of Insert is nil, it already is in the database
type Insert struct {
	*User
	*Device
	*Address
	*Identity
	*LogEntry

	UserHash     Hash
	DeviceHash   Hash
	AddressHash  Hash
	IdentityHash Hash

	UserID     int
	DeviceID   int
	AddressID  int
	IdentityID int
}

// DB represents a database
type DB struct {
	DB *sql.DB

	entries chan *Entry
	inserts chan *Insert

	queue map[int64]*Insert

	cache *Cache

	err error

	WriteInterval time.Duration
}

// Push passes the entry onto the queue to be processed
func (db *DB) Push(e *Entry) {
	db.entries <- e
}

// QueueLen returns then length of the writer queue
func (db *DB) QueueLen() int {
	return len(db.queue)
}

// worker processes entries from the queue and pushes them to the writer
func (db *DB) worker() {
	for {
		e := <-db.entries
		uH, dH, aH, iH := e.Hashes()

		var uID, dID, aID, iID int

		var u *User
		if uID = db.cache.Get(uH); uID == 0 {
			u = &User{
				ID:       uID,
				UID:      e.UID,
				Username: e.Username,
				FullName: e.FullName,
			}
		}
		var d *Device
		if dID = db.cache.Get(dH); dID == 0 {
			d = &Device{
				ID:               dID,
				Serial:           e.Serial,
				ClientIdentifier: e.ClientIdentifier,
				Hostname:         e.Hostname,
			}
		}
		var a *Address
		if aID = db.cache.Get(aH); aID == 0 {
			a = &Address{
				ID:         aID,
				IP:         e.IP,
				InternetIP: e.InternetIP,
			}
		}
		var i *Identity
		if iID = db.cache.Get(iH); iID == 0 {
			i = &Identity{
				ID:        iID,
				UserID:    uID,
				DeviceID:  dID,
				AddressID: aID,
			}
		}

		ins := &Insert{
			User:     u,
			Device:   d,
			Address:  a,
			Identity: i,

			LogEntry: &LogEntry{
				IdentityID: iID,
				Time:       e.Time,
			},

			UserHash:     uH,
			DeviceHash:   dH,
			AddressHash:  aH,
			IdentityHash: iH,

			UserID:     uID,
			DeviceID:   dID,
			AddressID:  aID,
			IdentityID: iID,
		}

		db.inserts <- ins
	}
}

// makeStmt creates a prepared statement or sets an error on db if one occurred
func makeStmt(db *DB, query string) *sql.Stmt {
	s, err := db.DB.Prepare(query)
	if err != nil {
		log.Println("Cannot create prepared statement:", query, "\n\t", err)
		db.err = err
	}
	return s
}

// getOrInsert gets the id of a row if it exists or creates the row and returns the new id
func getOrInsert(getStmt, insStmt *sql.Stmt, args ...interface{}) (id int, err error) {
	rID := new(int)

	row := getStmt.QueryRow(args...)
	err = row.Scan(rID)
	if err != nil && err != sql.ErrNoRows {
		logQueryError(getStmt, args...)
		return 0, err
	}
	if *rID != 0 {
		return *rID, nil
	}

	res, err := insStmt.Exec(args...)
	if err != nil {
		logQueryError(insStmt, args...)
		return 0, err
	}

	i, err := res.LastInsertId()
	if err != nil {
		logQueryError(insStmt, args...)
	}
	return int(i), err
}

// write polls the queue and every db.WriteInterval writes the data to the database and updates the cache
func (db *DB) writer() {
	var ins *Insert

	c := new(counter)

	lCache := NewCache()

mkstmts:
	stmts := map[string]*sql.Stmt{

		"uIns": makeStmt(db, "INSERT INTO user(uid, username, fullname) VALUES(?, ?, ?);"),
		"uGet": makeStmt(db, "SELECT id FROM user WHERE uid=? AND username=? AND fullname=?;"),

		"dIns": makeStmt(db, "INSERT INTO device(serial, clientidentifier, hostname) VALUES(?, ?, ?);"),
		"dGet": makeStmt(db, "SELECT id FROM device WHERE serial=? AND clientidentifier=? AND hostname=?;"),

		"aIns": makeStmt(db, "INSERT INTO address(ip, internetip) VALUES(?, ?);"),
		"aGet": makeStmt(db, "SELECT id FROM address WHERE ip=? AND internetip=?;"),

		"iIns": makeStmt(db, "INSERT INTO identity(user_id, device_id, address_id) VALUES(?, ?, ?);"),
		"iGet": makeStmt(db, "SELECT id FROM identity where user_id=? AND device_id=? AND address_id=?;"),

		"lIns": makeStmt(db, "INSERT INTO log(identity_id, time) VALUES(?, ?);"),
	}

	//if there is a problem getting to the database wait and try again
	if db.err != nil {
		db.err = nil
		time.Sleep(db.WriteInterval)
		goto mkstmts
	}

	timer := time.NewTimer(db.WriteInterval)

	for {

		select {
		case ins = <-db.inserts:
			db.queue[c.Next()] = ins
		case <-timer.C:
			timer.Reset(db.WriteInterval)

			if len(db.queue) == 0 {
				continue
			}

			log.Println("Inserting", len(db.queue), "entries")

			//start transaction
			tx, err := db.DB.Begin()
			if err != nil {
				log.Println("Error starting transaction:", err)
				continue
			}

			//make transaction version of stmt
			tstmts := make(map[string]*sql.Stmt)
			for k, v := range stmts {
				tstmts[k] = tx.Stmt(v)
			}

			//loop over queue
			for _, ins := range db.queue {

				//get or insert and get IDs
				if u := ins.User; u != nil {
					ins.UserID, err = getOrInsert(tstmts["uGet"], tstmts["uIns"], u.UID, u.Username, u.FullName)
					if err != nil {
						log.Println("Error getting or inserting user:", err)
						continue
					}
					lCache.Add(ins.UserHash, ins.UserID)
				}

				if d := ins.Device; d != nil {
					ins.DeviceID, err = getOrInsert(tstmts["dGet"], tstmts["dIns"], d.Serial, d.ClientIdentifier, d.Hostname)
					if err != nil {
						log.Println("Error getting or inserting device:", err)
						continue
					}
					lCache.Add(ins.DeviceHash, ins.DeviceID)
				}

				if a := ins.Address; a != nil {
					ins.AddressID, err = getOrInsert(tstmts["aGet"], tstmts["aIns"], a.IP, a.InternetIP)
					if err != nil {
						log.Println("Error getting or inserting address:", err)
						continue
					}
					lCache.Add(ins.AddressHash, ins.AddressID)
				}

				if i := ins.Identity; i != nil {
					ins.IdentityID, err = getOrInsert(tstmts["iGet"], tstmts["iIns"], ins.UserID, ins.DeviceID, ins.AddressID)
					if err != nil {
						log.Println("Error getting or inserting identity:", err)
						continue
					}
					lCache.Add(ins.IdentityHash, ins.IdentityID)
				}

				//insert LogEntry
				_, err = tstmts["lIns"].Exec(
					ins.IdentityID,
					ins.LogEntry.Time,
				)
				if err != nil {
					logQueryError(tstmts["lIns"], ins.IdentityID, ins.LogEntry.Time)
					log.Println("Error inserting log:", err)
				}
			} //end inner loop

			//commit
			err = tx.Commit()
			if err != nil {
				log.Println("Error commiting db:", err)
				continue
			}

			//clear queue
			for i := range db.queue {
				delete(db.queue, i)
			}

			//update db cache with entries then clear local cache
			lCache.Visit(func(key Hash, val int) {
				db.cache.Add(key, val)
			})
			lCache.Clear()

			//close stmts
			for _, v := range tstmts {
				err = v.Close()
				if err != nil {
					log.Println("Error closing statement:", err)
				}
			}
		}
	}
}

const queryLastUser = `
select 
	user.username,
	user.fullname,
    device.serial,
    address.ip,
    address.internetip,
	log.time
from (
    select
        max(log.id) as id
    from device
    inner join identity on device.id = identity.device_id
    inner join user on user.id = identity.user_id
    inner join log on log.identity_id = identity.id
    where
        device.serial in (%s) and
        user.username not in ('administrator', 'root', '')
    group by
        device.serial
) as joined
inner join log on
    joined.id = log.id
inner join identity on
    log.identity_id = identity.id
inner join device on
    identity.device_id = device.id
inner join user on
    identity.user_id = user.id
inner join address on
    identity.address_id = address.id
`

func (db *DB) QueryLastUser(serials []string) ([]*Entry, error) {
	params := make([]interface{}, len(serials))
	for idx, s := range serials {
		params[idx] = s
	}
	placeholder := strings.Join(strings.Split(strings.Repeat("?", len(serials)), ""), ",")
	query := fmt.Sprintf(queryLastUser, placeholder)

	rows, err := db.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("could not query db: %w", err)
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		e := new(Entry)
		if err := rows.Scan(&e.Username, &e.FullName, &e.Serial, &e.IP, &e.InternetIP, &e.Time); err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not scan rows: %w", err)
	}
	return nil, nil
}

// NewDB creates a new DB with the given driver and dsn as used by database/sql's Open.
// workers specifies how many worker goroutines will be used
func NewDB(driver, dsn string, workers int, writeInterval time.Duration) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	d := &DB{
		DB: db,

		entries: make(chan *Entry, workers*1000),
		inserts: make(chan *Insert, workers*1000),

		queue: make(map[int64]*Insert),

		cache: NewCache(),

		WriteInterval: writeInterval,
	}

	for i := 0; i < workers; i++ {
		go d.worker()
	}

	go d.writer()

	return d, nil
}
