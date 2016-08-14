package main

import (
    "github.com/stianeikeland/go-rpio"
    "os"
    "time"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "flag"
    "strconv"
)

const (
    InsertInitialSQL = "insert into last_motion(start_time) values (?)"
    InsertFinalSQL = "update last_motion set end_time=? where rid=?"
    SelectLastRidSQL = "select rid from last_motion order by rid desc limit 1"
)

var (
    motion_pin rpio.Pin
    light_pin rpio.Pin
    light_timeout int // this should be in seconds

    last_motion_ts time.Time
)

type MotionTracker struct {
    db *sql.DB

    last_motion_ts time.Time
    last_no_motion_ts time.Time

    rid int

    timeout int
}

func LogInfo(message string) {
    log.Println("INFO:", message)
}

func LogError(message string) {
    log.Println("ERROR:", message)
}

func LogFatal(err error) {
    log.Fatalln("FATAL:", err)
}

// Calculate if the 'to' - 'from' is > timeout (in seconds). Returns true of false
func MoreThanTimeout(from, to time.Time, timeout int) bool {
    return to.Sub(from).Seconds() > float64(timeout)
}

// get the latest rid in the db. this should be the id when data has just been inserted (i.e. no end_time yet)
func (mt *MotionTracker) GetSetDBRID() {
    var rid int
    mt.db.QueryRow(SelectLastRidSQL).Scan(&rid)

    mt.rid = rid
}

// insert a row into the db to start tracking motion. Inserts only the start_time and gets and sets the rid for this instance
// of MotionTracker
func (mt *MotionTracker) StartMotionRow() {
    stmt, err := mt.db.Prepare(InsertInitialSQL)
    if err != nil {
        LogFatal(err)
    } else {
        stmt.Exec(mt.last_motion_ts)
        mt.GetSetDBRID()
        LogInfo("New RID: " + strconv.Itoa(mt.rid))
    }
}

// Inserts the end_time for the rid of this instance of MotionTracker. Also sets the rid to `0` to signify that new motion
// can start being tracked
func (mt *MotionTracker) EndMotionRow() {
    stmt, err := mt.db.Prepare(InsertFinalSQL)
    if err != nil {
        LogFatal(err)
    } else {
        stmt.Exec(mt.last_no_motion_ts, mt.rid)
        mt.rid = 0
    }
}

// This should be called every time new motion is detected. If no current motion is being tracked, then it inserts a new row
// in the DB
func (mt *MotionTracker) TrackMotion() {
    mt.last_motion_ts = time.Now().UTC()
    if mt.rid == 0 {
        LogInfo("Starting logging to DB")
        mt.StartMotionRow()
    }
}

// This should be called every time no motion is detected. If motion was being tracked recently (i.e. no end_time has been recorded)
// then it'll record the end_time if the timeout has passed
func (mt *MotionTracker) TrackNoMotion() {
    mt.last_no_motion_ts = time.Now().UTC()
    if MoreThanTimeout(mt.last_motion_ts, mt.last_no_motion_ts, mt.timeout) && mt.rid != 0 {
        LogInfo("Ending logging to DB")
        mt.EndMotionRow()
    }
}

func NewMotionTracker(timeout int) *MotionTracker {
    db, err := sql.Open("sqlite3", "last_motion.db")
    if err != nil {
        LogFatal(err)
    }

    return &MotionTracker{db: db, timeout: timeout}
}

func TurnLightOn() {
    if light_pin.Read() != rpio.Low {
        LogInfo("Turning light on")
        light_pin.Low()
    }
}

func TurnLightOff() {
    if light_pin.Read() != rpio.High {
        LogInfo("Turning light off")
        light_pin.High()
    }
}

func main() {
    var mpin, lpin int
    var logfile_location string

    flag.IntVar(&mpin, "motion_pin", 7, "Pin for the motion detector (BCM mode). Defaults to 7")
    flag.IntVar(&lpin, "light_pin", 11, "Pin for the light (BCM mode). Defaults to 11")
    flag.IntVar(&light_timeout, "timeout", 180, "Timeout before turning off the light. Value should be in seconds. Defaults to 180")
    flag.StringVar(&logfile_location, "logfile", "/var/log/motion.log", "Location for the logfile. Defaults to /var/log/motion.log")

    flag.Parse()

    f, err := os.OpenFile(logfile_location, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        LogFatal(err)
    }
    defer f.Close()

    log.SetOutput(f)

    LogInfo("Using motion pin: " + strconv.Itoa(mpin))
    LogInfo("Using light pin: " + strconv.Itoa(lpin))
    LogInfo("Using timeout: " + strconv.Itoa(light_timeout))
    LogInfo("Using logfile: " + logfile_location)

    motion_pin = rpio.Pin(mpin)
    light_pin = rpio.Pin(lpin)

    if err := rpio.Open(); err != nil {
        LogFatal(err)
    }

    mt := NewMotionTracker(light_timeout)

    defer rpio.Close()

    motion_pin.Input()
    light_pin.Output()

    for true {
        switch motion_pin.Read() {
        case rpio.High: // this is the one which means motion was detected
            last_motion_ts = time.Now().UTC()
            mt.TrackMotion()
        case rpio.Low:
            mt.TrackNoMotion()
        }

        if !MoreThanTimeout(last_motion_ts, time.Now().UTC(), light_timeout) {
            TurnLightOn()
        } else {
            TurnLightOff()
        }

        // sleep for 100 ms
        time.Sleep(time.Millisecond * 100)
    }
}
