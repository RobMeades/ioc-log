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
    "bytes"
    "strings"
    "time"
    "encoding/binary"
)

//--------------------------------------------------------------------
// Types
//--------------------------------------------------------------------

// Struct to hold a log item
type LogItem struct {
    Timestamp   int // in uSeconds
    Enum        int
    Parameter   int
}

//--------------------------------------------------------------------
// Constants
//--------------------------------------------------------------------

// The size of a single log entry in bytes
const LOG_ITEM_SIZE int = 12

// A magic enum item, in which the parameter is the UTC time
const LOG_ITEM_UTC_TIME int = 2

//--------------------------------------------------------------------
// Variables
//--------------------------------------------------------------------

// The base time of a log
var logTimeBase int64

// The timestamp at the base time above
var logTimestampAtBase int64

//--------------------------------------------------------------------
// Functions
//--------------------------------------------------------------------

// Open a log file
func openLogFile(directory string, clientIpAddress string) *os.File {
    // File name is the IP address of the client (port number removed),
    // the dots replaced with dashes, followed by the UTC time so:
    // 154-46-789-1_2017-11-17_15-35-01.log
    fileName := fmt.Sprintf("%s%c%s_%s.log", directory, os.PathSeparator, strings.Replace(strings.Split(clientIpAddress, ":")[0], ".", "-", -1), time.Now().UTC().Format("2006-01-02_15-04-05"))
    logFile, err := os.Create(fileName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating logfile (%s).\n", err.Error())
    }
    
    return logFile 
}

// Handle a log item
func handleLogItem(itemIn []byte, logFile *os.File) {
    var item LogItem
    var err error
    var itemString string
    var enumString string
    
    if len(itemIn) >= LOG_ITEM_SIZE {
        buf := bytes.NewReader(itemIn)
        err = binary.Read(buf, binary.LittleEndian, &item.Timestamp)
        if err != nil {
            err = binary.Read(buf, binary.LittleEndian, &item.Enum)
        } 
        if err != nil {
            err = binary.Read(buf, binary.LittleEndian, &item.Parameter)
        }        
        if err != nil {
            // We have a log item, translate it to text
            if item.Enum < int(C.gNumLogStrings) {
                enumString = C.GoString((*C.char) (unsafe.Pointer(uintptr(unsafe.Pointer(C.gLogStrings)) + uintptr(item.Enum) * unsafe.Sizeof(C.gLogStrings))))
                if item.Enum == LOG_ITEM_UTC_TIME {
                    logTimeBase = int64(item.Parameter)
                    logTimestampAtBase = int64(item.Timestamp)
                }
            } else {
                enumString = fmt.Sprintf("unknown (%#x)", C.gNumLogStrings)
            }
            microsecondTime := time.Unix(logTimeBase, (int64(item.Timestamp) - logTimestampAtBase) * 1000).UTC()
            nanosecondTime := microsecondTime.Nanosecond()
            timeString := fmt.Sprintf("%s_%03d.%03d", microsecondTime.Format("2006-01-02_15-04-05"), nanosecondTime / 1000000, nanosecondTime / 1000)
            itemString = fmt.Sprintf("%s: %s %d (%#x)\n", timeString, enumString, item.Parameter, item.Parameter)
        }
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
                logTimeBase = 0;
                logTimestampAtBase = 0;
                                
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
    loggingServer(port, directory)
}
