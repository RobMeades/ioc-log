/* Logging input server for the Internet of Chuffs.
 *
 * Copyright (C) u-blox Melbourn Ltd
 * u-blox Melbourn Ltd, Melbourn, UK
 *
 * All rights reserved.
 *
 * This source file is the sole property of u-blox Melbourn Ltd.
 * Reproduction or utilization of this source in whole or part is
 * forbidden without the written consent of u-blox Melbourn Ltd.
 */

package main

/*
#cgo CFLAGS: -I.

#include "log_enum.h"
extern const char *gLogStrings;
extern const int gNumLogStrings;
*/
import "C"

import (
	"unsafe"
    "fmt"
    "net"
    "os"
    "log"
    "strings"
    "time"
    "encoding/binary"
)

//--------------------------------------------------------------------
// Types
//--------------------------------------------------------------------

// Struct to hold a log item
type LogItem struct {
    Timestamp   uint32 // in uSeconds
    Enum        uint32
    Parameter   uint32
}

//--------------------------------------------------------------------
// Constants
//--------------------------------------------------------------------

// The size of a single log entry in bytes
const LOG_ITEM_SIZE int = 12

// A minimum value for current Unix time
const UNIX_TIME_MIN uint32 = 1510827960

//--------------------------------------------------------------------
// Variables
//--------------------------------------------------------------------

// The base time of a log
var logTimeBase int64

// The timestamp at the base time above
var logTimestampAtBase int64

// The array of C strings as a slice
var cLogStrings []*C.char

//--------------------------------------------------------------------
// Functions
//--------------------------------------------------------------------

// Open a log file
func openLogFile(directory string, clientIpAddress string) *os.File {
    // File name is the IP address of the client (port number removed),
    // the dots replaced with dashes, followed by the UTC time, with
    // an x at the start to indicate that this is currently server time
    // so: x154-46-789-1_2017-11-17_15-35-01.log
    fileName := fmt.Sprintf("%s%cx%s_%s.log", directory, os.PathSeparator, strings.Replace(strings.Split(clientIpAddress, ":")[0], ".", "-", -1), time.Now().UTC().Format("2006-01-02_15-04-05"))
    logFile, err := os.Create(fileName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating logfile (%s).\n", err.Error())
    }
    
    return logFile 
}

// Rename a log file
func renameLogFile(oldName string, logTime int64) bool {    
    // Find the _ in the old name and replace the bits after it
    // with a time generated from logTime
    newName := fmt.Sprintf("%s_%s.log", strings.Split(oldName, "_")[0], time.Unix(logTime, 0).UTC().Format("2006-01-02_15-04-05"))
    err := os.Rename(oldName, newName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error renaming logfile from \"%s\" to \"%s\" (%s).\n", oldName, newName, err.Error())
    }
    
    return (err == nil); 
}

// Handle a log item
func handleLogItem(itemIn []byte, logFile *os.File) {
    var item LogItem
    var itemString string
    var enumString string
    
    if len(itemIn) >= LOG_ITEM_SIZE {
        item.Timestamp = binary.LittleEndian.Uint32(itemIn[0:])
        item.Enum = binary.LittleEndian.Uint32(itemIn[4:])
        item.Parameter = binary.LittleEndian.Uint32(itemIn[8:])
        // We have a log item, translate it to text
        if item.Enum < uint32(C.gNumLogStrings) {
            enumString = C.GoString(cLogStrings[item.Enum])
            // If a current time marker arrives and it looks real then grab it and use it from now on
            if (item.Enum == C.EVENT_CURRENT_TIME_UTC) && (item.Parameter > UNIX_TIME_MIN) && (logTimeBase == 0) {
                logTimeBase = int64(item.Parameter)
                logTimestampAtBase = int64(item.Timestamp)
            }
        } else {
            enumString = fmt.Sprintf("unknown (%#x)", item.Enum)
        }
        microsecondTime := time.Unix(logTimeBase, (int64(item.Timestamp) - logTimestampAtBase) * 1000).UTC()
        nanosecondTime := microsecondTime.Nanosecond()
        timeString := fmt.Sprintf("%s_%03.3f", microsecondTime.Format("2006-01-02_15-04-05"), float64(nanosecondTime / 1000) / 1000)
        itemString = fmt.Sprintf("%s: %s [%d] %d (%#x)\n", timeString, enumString, item.Enum, item.Parameter, item.Parameter)
    }
    
    if (logFile != nil) && (itemString != "") {
        logFile.WriteString(itemString)
    }
}

// Run a TCP server forever
func loggingServer(port string, directory string) {
    var logFile *os.File
    var newServer net.Conn
    var currentServer net.Conn
    
    listener, err := net.Listen("tcp", ":" + port)
    if err == nil {
        defer listener.Close()
        // Listen for a connection
        for {
            fmt.Printf("Logging server waiting for a [further] TCP connection on port %s.\n", port)    
            newServer, err = listener.Accept()
            if err == nil {
                if currentServer != nil {
                    currentServer.Close()
                }
                if logFile != nil {
                    logFile.Close()
                    if logTimeBase != 0 {
                        renameLogFile(logFile.Name(), logTimeBase)
                        logTimeBase = 0;
                        logTimestampAtBase = 0;
                    }
                }
                currentServer = newServer
                x, success := currentServer.(*net.TCPConn)
                if success {
                    err1 := x.SetNoDelay(true)
                    if err1 != nil {
                        log.Printf("Unable to switch off Nagle algorithm (%s).\n", err1.Error())
                    }
                } else {
                    log.Printf("Can't cast *net.Conn to *net.TCPConn in order to configure it.\n")
                }
                fmt.Printf("Logging connection made by %s.\n", currentServer.RemoteAddr().String())
                logFile = openLogFile(directory, currentServer.RemoteAddr().String())
                                
                if logFile != nil {
                    // Process datagrams received items in a go routine
                    go func(server net.Conn) {
                        // Read log items until the connection is closed under us
                        line := make([]byte, LOG_ITEM_SIZE)                
                        for numBytesIn, err := server.Read(line); (err == nil) && (numBytesIn > 0); numBytesIn, err = server.Read(line) {
                            handleLogItem(line[:numBytesIn], logFile)
                        }
                        fmt.Printf("[Logging connection to %s closed].\n", server.RemoteAddr().String())
                    }(currentServer)
                }                
            } else {
                fmt.Fprintf(os.Stderr, "Error accepting logging connection (%s).\n", err.Error())
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "Unable to listen for logging connections on port %s (%s).\n", port, err.Error())        
    }
}

// Run the logging input server; this function should never return
func operateLoggingInputServer(port string, directory string) {
    // Convert the array of C strings into a slice to make it possible to index into it
    cLogStrings = (*[1 << 30]*C.char)(unsafe.Pointer(&C.gLogStrings))[:C.gNumLogStrings]
    
    loggingServer(port, directory)
}
