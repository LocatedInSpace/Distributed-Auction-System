package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	DAS "github.com/LocatedInSpace/Distributed-Auction-System/proto"
	"google.golang.org/grpc"
)

const BASEPORT = 7000         // port offset to start servers from
const DELAYED_MUTEX = true    // this setting makes it way more likely for servers to stay in sync
const PRECISE_LOGGING = false // ups precision on timestamps

type Replica struct {
	DAS.UnimplementedDASServer
	port     uint16     // used for logging
	mutex    sync.Mutex // used to lock the server to avoid race conditions.
	auctions []Auction
}

type Auction struct {
	highestBid   uint64
	bidder       uint32
	item         string
	auctionStart time.Time
	duration     uint32
}

func main() {
	var port uint16 = BASEPORT
	started := false
	var list net.Listener
	var err error
	// loop for starting server up on next free port (BASEPORT++)
	for !started {
		list, err = net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
		if err != nil {
			log.Printf("Could not open listener on port %v, retrying port++\n", port)
			port++
		} else {
			started = true
		}
	}
	log.Printf("Created listener on port %v\n", port)

	f := setLog(port)
	defer f.Close()

	grpcServer := grpc.NewServer()

	server := &Replica{
		port: port,
	}

	DAS.RegisterDASServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Replica started on %v\n", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

// func (r *Replica) Bid(ctx context.Context, amount *DAS.Amount) (*DAS.Ack, error) {
func (r *Replica) Bid(ctx context.Context, amount *DAS.Amount) (*DAS.Ack, error) {
	r.mutex.Lock()
	log.Printf("Bid() | Request received from %v, amount: %v\n", amount.Id, amount.Bid)
	// no active auctions
	if len(r.auctions) == 0 {
		log.Printf("Bid() | Told %v, no active auctions\n", amount.Id)
		if DELAYED_MUTEX {
			go r.DelayedUnlock()
		} else {
			r.mutex.Unlock()
		}
		return &DAS.Ack{
			Response: DAS.Acks_EXCEPTION,
			Message:  "No active auction to bid on",
		}, nil
	} else {
		// notice we use a reference, which means changes to lastAuction get "saved"
		lastAuction := &r.auctions[len(r.auctions)-1]
		now := time.Now()
		difference := now.Sub(lastAuction.auctionStart)
		// last auction is over
		if difference.Milliseconds() > int64(lastAuction.duration) {
			log.Printf("Bid() | Told %v, auction is over\n", amount.Id)
			if DELAYED_MUTEX {
				go r.DelayedUnlock()
			} else {
				r.mutex.Unlock()
			}
			return &DAS.Ack{
				Response: DAS.Acks_EXCEPTION,
				Message:  "Auction is over",
			}, nil
		} else {
			if amount.Bid > lastAuction.highestBid {
				lastAuction.bidder = amount.Id
				lastAuction.highestBid = amount.Bid
				log.Printf("Bid() | Accepted bid from %v\n", amount.Id)
				if DELAYED_MUTEX {
					go r.DelayedUnlock()
				} else {
					r.mutex.Unlock()
				}
				return &DAS.Ack{
					Response: DAS.Acks_SUCCESS,
					Message:  "Bid increased",
				}, nil
			} else {
				log.Printf("Bid() | Rejected bid from %v\n", amount.Id)
				if DELAYED_MUTEX {
					go r.DelayedUnlock()
				} else {
					r.mutex.Unlock()
				}
				return &DAS.Ack{
					Response: DAS.Acks_FAIL,
					Message:  "Bid is lower than the highest bid",
				}, nil
			}
		}
	}
}

func (r *Replica) Result(ctx context.Context, _ *DAS.Empty) (*DAS.Outcome, error) {
	r.mutex.Lock()
	// there exist no auctions, so return empty outcome
	if len(r.auctions) == 0 {
		log.Printf("Result() | Told client that there have been no auctions\n")
		if DELAYED_MUTEX {
			go r.DelayedUnlock()
		} else {
			r.mutex.Unlock()
		}
		return &DAS.Outcome{}, nil
	} else {
		lastAuction := r.auctions[len(r.auctions)-1]
		now := time.Now()
		difference := now.Sub(lastAuction.auctionStart)
		var left uint32
		// last auction is over
		if difference.Milliseconds() >= int64(lastAuction.duration) {
			log.Printf("Result() | Sent last auction, '%s' lasted %vms, won by id %v\n", lastAuction.item, lastAuction.duration, lastAuction.bidder)
			left = 0
		} else {
			log.Printf("Result() | Sent current auction, '%s' lasts %vms, id %v is winning\n", lastAuction.item, lastAuction.duration, lastAuction.bidder)
			left = lastAuction.duration - uint32(difference.Milliseconds())
		}

		if DELAYED_MUTEX {
			go r.DelayedUnlock()
		} else {
			r.mutex.Unlock()
		}
		return &DAS.Outcome{
			Left:   left,
			Amount: lastAuction.highestBid,
			Bidder: lastAuction.bidder,
			Item:   lastAuction.item,
		}, nil
	}
}

func (r *Replica) StartAuction(ctx context.Context, item *DAS.Item) (*DAS.Ack, error) {
	r.mutex.Lock()
	// there exist no auctions, no need to check if last one is active
	if len(r.auctions) == 0 {
		log.Printf("Auction() | Started auction '%v', duration: %v\n", item.Name, item.Alive)
		r.auctions = append(r.auctions,
			Auction{
				highestBid:   item.Start,
				bidder:       0,
				item:         item.Name,
				auctionStart: time.Now(),
				duration:     item.Alive,
			})
	} else {
		lastAuction := r.auctions[len(r.auctions)-1]
		now := time.Now()
		difference := now.Sub(lastAuction.auctionStart)
		// last auction is over, so we just append this as current auction
		if difference.Milliseconds() > int64(lastAuction.duration) {
			log.Printf("Auction() | Started auction '%v', duration: %v\n", item.Name, item.Alive)
			r.auctions = append(r.auctions,
				Auction{
					highestBid:   item.Start,
					bidder:       0,
					item:         item.Name,
					auctionStart: time.Now(),
					duration:     item.Alive,
				})
		} else {
			log.Printf("Auction() | Rejected auction '%v', '%v' is currently live\n", item.Name, lastAuction.item)
			if DELAYED_MUTEX {
				go r.DelayedUnlock()
			} else {
				r.mutex.Unlock()
			}
			return &DAS.Ack{
				Response: DAS.Acks_FAIL,
				Message:  "An auction is already running",
			}, nil
		}
	}
	if DELAYED_MUTEX {
		go r.DelayedUnlock()
	} else {
		r.mutex.Unlock()
	}
	return &DAS.Ack{
		Response: DAS.Acks_SUCCESS,
	}, nil
}

func (r *Replica) Ping(ctx context.Context, _ *DAS.Empty) (*DAS.Empty, error) {
	return &DAS.Empty{}, nil
}

func (r *Replica) DelayedUnlock() {
	time.Sleep(5 * time.Millisecond)
	r.mutex.Unlock()
}

// sets the logger to use a log.txt file instead of the console
func setLog(port uint16) *os.File {
	filename := fmt.Sprintf("replica-%v.txt", port)
	// Clears the log.txt file when a new server is started
	if err := os.Truncate(filename, 0); err != nil {
		log.Printf("Failed to truncate: %v\n", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	// print to both file and console
	mw := io.MultiWriter(os.Stdout, f)

	if PRECISE_LOGGING {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}
	log.SetOutput(mw)
	return f
}
