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

import (
    "crypto/tls"
    "fmt"
    "net"
    "os"
    "time"
    "encoding/hex"
)

type Proxy struct {
    Target *net.TCPAddr
    Addr *net.TCPAddr
    Config *tls.Config
    Timeout time.Duration
}

func NewProxy(target *net.TCPAddr, config *tls.Config) *Proxy {
    p := &Proxy{
        Target:   target,
        Timeout:  time.Minute,
        Config:   config,
    }
    return p
}

func (p *Proxy) ListenAndServe(laddr *net.TCPAddr) {
    p.Addr = laddr

    var listener net.Listener
    listener, err := net.ListenTCP("tcp", laddr)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    p.serve(listener)
}

func (p *Proxy) ListenAndServeTLS(laddr *net.TCPAddr, certFile, keyFile string) {
    p.Addr = laddr

    var listener net.Listener
    cer, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        fmt.Println(err)
        return
    }
    config := &tls.Config{Certificates: []tls.Certificate{cer}}
    listener, err = tls.Listen("tcp", laddr.String(), config)
    if err != nil {
        fmt.Println(err)
        return
    }

    p.serve(listener)
}

func (p *Proxy) serve(ln net.Listener) {
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }

        go p.handleConn(conn)
    }
}

func closeConnection(conn net.Conn){
    fmt.Println("Close connection with ", conn.RemoteAddr().String())
    conn.Close()
}

func (p *Proxy) handleConn(conn net.Conn) {
    var rconn net.Conn
    var err error
    if p.Config == nil {
        rconn, err = net.Dial("tcp", p.Target.String())
    } else {
        rconn, err = tls.Dial("tcp", p.Target.String(), p.Config)
    }
    if err != nil {
        fmt.Println(err)
        return
    }


    pipe(conn, rconn)
}


func pipe(conn1 net.Conn, conn2 net.Conn) {
    chan1 := chanFromConn(conn1)
    chan2 := chanFromConn(conn2)

    var buffer []byte = nil
    var t time.Time
    var way int

    way = 0
    

    for {
        select {
        case b1 := <-chan1:
            if b1 == nil {
                flushBuffer(&buffer, way, t, conn1, conn2)
                closeConnection(conn1)
                closeConnection(conn2)
                closeChans(chan1, chan2)
                return
            } else {
                switch way {
                    case 0, 1:
                        buffer = append(buffer, b1...)
                        t = time.Now()
                    case 2:
                        fmt.Println(t.Format("2006-01-02 15:04:05"), " : ", conn2.RemoteAddr().String(), " -> ", conn1.RemoteAddr().String())
                        fmt.Println(hex.Dump(buffer))
                        buffer = nil
                        buffer = append(buffer, b1...)
                }
                way = 1
                conn2.Write(b1)
            }
        case b2 := <-chan2:
            if b2 == nil {
                flushBuffer(&buffer, way, t, conn1, conn2)
                closeConnection(conn1)
                closeConnection(conn2)
                closeChans(chan1, chan2)
                return
            } else {
                switch way {
                    case 0, 2:
                        buffer = append(buffer, b2...)
                        t = time.Now()
                    case 1:
                        fmt.Println(t.Format("2006-01-02 15:04:05"), " : ", conn1.RemoteAddr().String(), " -> ", conn2.RemoteAddr().String())
                        fmt.Println(hex.Dump(buffer))
                        buffer = nil
                        buffer = append(buffer, b2...)
                }
                way = 2
                conn1.Write(b2)
            }
        case <-time.After(500 * time.Millisecond):
            flushBuffer(&buffer, way, t, conn1, conn2)
        }
    }
}

func closeChans(c1, c2 chan []byte){
    for {
        select {
            case <-c1:
            case <-c2:
            case <-time.After(5 * time.Millisecond):
                return
        }
    }
}

func flushBuffer(buffer * []byte, way int, t time.Time, conn1, conn2 net.Conn){
    if len(*buffer) > 0 {
        switch way {
        case 1:
            fmt.Println(t.Format("2006-01-02 15:04:05"), " : ", conn1.RemoteAddr().String(), " -> ", conn2.RemoteAddr().String())
            fmt.Println(hex.Dump(*buffer))
            *buffer = nil
        case 2:
            fmt.Println(t.Format("2006-01-02 15:04:05"), " : ", conn2.RemoteAddr().String(), " -> ", conn1.RemoteAddr().String())
            fmt.Println(hex.Dump(*buffer))
            *buffer = nil
        }
    }
}


func chanFromConn(conn net.Conn) chan []byte {
    c := make(chan []byte)
    go func() {
        b := make([]byte, 4096)

        for {
            n, err := conn.Read(b)
            if n > 0 {
                res := make([]byte, n)
                copy(res, b[:n])
                c <- res
            }
            if err != nil {
                c <- nil
                break
            }
        }
    }()
    return c
}




