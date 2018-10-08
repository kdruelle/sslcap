/*******************************************************************************
 **
 ** MIT License
 **
 ** Copyright (c) 2017 Kevin Druelle
 **
 ** Permission is hereby granted, free of charge, to any person obtaining a copy
 ** of this software and associated documentation files (the "Software"), to deal
 ** in the Software without restriction, including without limitation the rights
 ** to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 ** copies of the Software, and to permit persons to whom the Software is
 ** furnished to do so, subject to the following conditions:
 **
 ** The above copyright notice and this permission notice shall be included in all
 ** copies or substantial portions of the Software.
 **
 ** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 ** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 ** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 ** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 ** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 ** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 ** SOFTWARE
 **
 ******************************************************************************/

package main

import(
    "os"
    "fmt"
    "net"
    "runtime"
    "crypto/tls"
    "time"
    "github.com/spf13/cobra"
)

var (
    help       bool
    version    bool
    localAddr  string
    remoteAddr string
    localCert  string
    localKey   string
    remoteTLS  bool
    timeout    int
    versionStr = "undefined"
    buildTime  = "undefined"
)

func init(){
    //cobra.OnInitialize();
    rootCmd.AddCommand(versionCmd);
    rootCmd.PersistentFlags().BoolVarP(&help, "help", "h", false, "Print this help message")
    rootCmd.PersistentFlags().BoolVarP(&version, "version", "v", false, "Print version informations")
    rootCmd.PersistentFlags().StringVarP(&localAddr, "laddr", "l", ":4444", "proxy local address")
    rootCmd.PersistentFlags().StringVarP(&remoteAddr, "raddr", "r", ":80", "proxy remote address")
    rootCmd.PersistentFlags().StringVarP(&localCert, "lcert", "c", "", "proxy certificate x509 file for tls/ssl use")
    rootCmd.PersistentFlags().StringVarP(&localKey, "lkey", "k", "", "proxy key x509 file for tls/ssl use")
    rootCmd.PersistentFlags().BoolVarP(&remoteTLS, "rtls", "t", false, "tls/ssl between proxy and target")
    rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "u", 0, "wait  seconds before closing second pipe")
}


var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version information and exits",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("SSLCap %s    %s/%s    %s\n", versionStr, runtime.GOOS, runtime.GOARCH, buildTime)
    },
}



var rootCmd = &cobra.Command{
    Use:   "sslcap",
    Short: "Simple TCP Proxy",
    Long: `SSLcap is a simple TCP Proxy supporting SSL/TLS`,
    Run: func(cmd *cobra.Command, args []string) {
        if help {
            cmd.Usage()
            return
        }

        if version {
            fmt.Println("Simple TCP Proxy version 1.0")
            return
        }

        laddr, err := net.ResolveTCPAddr("tcp", localAddr)
        if err != nil {
            fmt.Println(err.Error())
            os.Exit(1)
        }

        raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
        if err != nil {
            fmt.Println(err.Error())
            os.Exit(1)
        }

        localTLS := false
        if localCert != "" || localKey != "" {
            localTLS = true
            if localCert == "" || localKey == "" {
                fmt.Println("Certificate and key file required")
                os.Exit(1)
            }
            if !exists(localCert) {
                fmt.Println("Unable to load certificate file ", localCert)
                os.Exit(1)
            }
            if !exists(localKey) {
                fmt.Println("Unable to load key file ", localKey)
                os.Exit(1)
            }
        }

        var p = new(Proxy)
        if remoteTLS {
            p = NewProxy(raddr, &tls.Config{InsecureSkipVerify: true})
        } else {
            p = NewProxy(raddr, nil)
        }

        p.Timeout = time.Duration(timeout) * time.Second

        fmt.Println("Proxying from " + laddr.String() + " to " + p.Target.String())
        if localTLS {
            p.ListenAndServeTLS(laddr, localCert, localKey)
        } else {
            p.ListenAndServe(laddr)
        }
    },
}


func main(){
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }
}

func exists(filename string) bool {
    _, err := os.Stat(filename)
    return !os.IsNotExist(err)
}
