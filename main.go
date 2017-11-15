/* main() for Internet of Chuffs logging server.
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

import (
    "fmt"
    "os"
    "github.com/jessevdk/go-flags"
)

// This is the Internet of Chuffs logging server.
// It accepts binary logging from the Internet of Chuffs
// client on the given port and writes the logging to a
// text file for easy reading.

//--------------------------------------------------------------------
// Constants
//--------------------------------------------------------------------

//--------------------------------------------------------------------
// Variables
//--------------------------------------------------------------------

// Command-line items
var opts struct {
    Required struct {
        InPort string `positional-arg-name:"input-port" description:"the input port for incoming binary logging"`
    } `positional-args:"true" required:"yes"`
    LogDir string `short:"o" long:"logdir" default:"." description:"the directory in which to place log files"`
}

//--------------------------------------------------------------------
// Functions
//--------------------------------------------------------------------

// Deal with command-line parameters
func cli() {
     _, err := flags.Parse(&opts)

    if err != nil {
        os.Exit(-1)        
    }    
}

// Entry point
func main() {
    var err error;
    
    // Handle the command line
    cli()
    
    // Make sure the directory path exists
    err = os.MkdirAll(opts.LogDir, os.ModePerm);
    if (err != nil) {
        fmt.Fprintf(os.Stderr, "Unable to create directory %s for log files (%s).\n", opts.LogDir, err.Error())
        os.Exit(-1)
    } else {
        fmt.Printf("Log files will be written to the directory \"%s\".\n", opts.LogDir)
    }
    
    // Run the incoming logging server loop (which should block)
    operateLoggingInputServer(opts.Required.InPort, opts.LogDir)
}
