package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	DAS "github.com/LocatedInSpace/Distributed-Auction-System/proto"

	"google.golang.org/grpc"
)

const BASEPORT = 7000 // port offset to look for servers from
const REPLICAS = 5    // amount of replicas we've started up

const VERBOSE = false // print each response from each replica

const AUTOCLIENT = true // will randomly call startauction, sendbids, etc.
// this parameter was used to generate the logs that verify replicas are in sync

const MIN_DELAY = 5  // mindelay before next AUTOCLIENT roll
const MAX_DELAY = 30 // maxdelay before next AUTOCLIENT roll
type ReplicaServers struct {
	clients []DAS.DASClient
	ctx     context.Context
}

var clientToPort map[DAS.DASClient]int32

var id uint32

func main() {
	idUint64, err := strconv.ParseUint(os.Args[1], 10, 32)
	if err != nil || idUint64 == 0 {
		log.Fatalf("You need to supply a valid uint32 value > 0")
	}
	id = uint32(idUint64)

	clientToPort = make(map[DAS.DASClient]int32)

	f := setLog(id)
	defer f.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// connect to all replicas
	//var servers [REPLICAS]DAS.DASClient
	server := &ReplicaServers{
		ctx: ctx,
	}

	var allReplicasDead bool = true
	for i := 0; i < REPLICAS; i++ {
		port := BASEPORT + int32(i)

		var conn *grpc.ClientConn
		// 1 second timeout
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		conn, err := grpc.DialContext(ctx, fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Printf("Dial (port %v) failed: %s", port, err)
			continue
		} else {
			log.Printf("Dial (port: %v) succeeded\n", port)
		}
		defer conn.Close()

		allReplicasDead = false
		c := DAS.NewDASClient(conn)
		server.clients = append(server.clients, c)
		clientToPort[c] = port
	}
	if allReplicasDead {
		log.Fatalf("Could not find any replicas - are you sure they are running?")
	}

	if AUTOCLIENT {
		rand.Seed(time.Now().UnixNano())
		item := fmt.Sprintf("item-%v", id)
		var action int
		for {
			action = rand.Intn(10)
			server.PurgeDeadReplicas()
			switch action {
			case 0:
				fmt.Println(server.SendBid(server.GetResults().Amount + 1))
			case 1:
				server.SendBid(rand.Uint64())
			case 2:
				fmt.Println(FormatOutcome(server.GetResults()))
			case 3:
				fmt.Println(server.StartAuction(rand.Uint64(), uint32(rand.Intn(65535)), item))
			}
			// 1e6 = millisecond
			time.Sleep(time.Duration((MIN_DELAY + rand.Intn(MAX_DELAY-MIN_DELAY)) * 1e6))
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("-- Enter 'h' for help --")
		for {
			fmt.Print("-> ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if len(text) == 0 {
				continue
			}

			input := strings.Fields(text)

			server.PurgeDeadReplicas()
			if input[0] == "h" {
				fmt.Println(`| 'h' displays commands & their syntax
| 'b *amount' bids on auction, with * being a number
|     if amount is empty, then we assume that we want to increment bid by 1
| 'r' gets the result of the active (or last) auction
| 's *start *duration *name' starts an auction lasting duration, for item with name, & starting bid`)
			} else if input[0] == "b" {
				if len(input) == 1 {
					// get highest bid - add one
					log.Println(server.SendBid(server.GetResults().Amount + 1))
				} else {
					bid, err := strconv.ParseUint(input[1], 10, 64)
					if err != nil {
						fmt.Println("The second parameter of 'b' MUST be a uint64")
						continue
					}
					log.Println(server.SendBid(bid))
				}
			} else if input[0] == "r" {
				log.Println(FormatOutcome(server.GetResults()))
			} else if input[0] == "s" {
				name := ""
				if len(input) < 4 {
					fmt.Println("Missing parameters - 4 are expected")
					continue
				} else {
					name = strings.Join(input[3:], " ")
				}

				start, err := strconv.ParseUint(input[1], 10, 64)
				if err != nil {
					fmt.Println("The second parameter of 's' MUST be a uint64")
					continue
				}

				duration, err := strconv.ParseUint(input[2], 10, 32)
				if err != nil {
					fmt.Println("The third parameter of 's' MUST be a uint32")
					continue
				}
				log.Println(server.StartAuction(start, uint32(duration), name))
			} else {
				fmt.Println("Command not recognized :(")
			}
		}
	}
}

func (s *ReplicaServers) PurgeDeadReplicas() {
	// why is this needed? since we are removing old replicas in sendbid & startauction - lets imagine the following situation:
	// client A has called result - finds out some servers are dead
	// then removes them from own replicaservers
	// now client B (who still has an old replicaserver) tries to start an auction
	// client B will timeout on the dead servers for one second - in one of these timeouts, client A starts an auction but will incur no timeouts
	// the servers are now out of sync - client A will reach the later servers, before client B, even though client B called first

	// ok - so why not NOT purge servers then? while this *can* be fine, there is a chance of the following situation:
	// client A has a dead replica in replicaservers, the second to last one. there is one second left in the auction
	// client A sends a bid, replicaserver 1..N-2 all accept it - replicaserver N-1 is timed out, client A waits a second here
	// replicaserver N declines client A bid, since it is after the auction has ended.

	// passive replication would definitely be a better solution for auction, since we only have to worry about central replica time then
	// and technically since the servers will be a few milliseconds out of sync - there is the *slightest* chance of above situation happening without
	// timeouts - even though it is incredibly slim (given the assignment assumption of transmissions happening with a known time-limit, this is combated by using the servers local time as timestamp, however
	// this is only a theoretically correct implementatiom in practice, there will be differences in how long a transmission takes on the same network under the same conditions)
	query := &DAS.Empty{}

	var remove []int
	for i, r := range s.clients {
		_, err := r.Ping(s.ctx, query)
		if err != nil {
			remove = append(remove, i)
			continue
		}
	}

	for i, val := range remove {
		// remove from ReplicaServers clients these indexes - since calling them failed
		s.clients = append(s.clients[:val-i], s.clients[(val-i)+1:]...)
	}
}

func (s *ReplicaServers) SendBid(amount uint64) *DAS.Ack {
	query := &DAS.Amount{
		Id:  id,
		Bid: amount,
	}

	var responses []*DAS.Ack
	var remove []int
	if VERBOSE {
		fmt.Println("--- All responses ---")
	}
	for i, r := range s.clients {
		ack, err := r.Bid(s.ctx, query)
		if err != nil {
			remove = append(remove, i)
			if VERBOSE {
				log.Printf("Port %v | %s\n", clientToPort[r], err)
			}
			continue
		}
		if VERBOSE {
			log.Printf("Port %v | %s\n", clientToPort[r], ack)
		}
		responses = append(responses, ack)
	}
	if VERBOSE {
		fmt.Println("---------------------")
	}

	for i, val := range remove {
		// remove from ReplicaServers clients these indexes - since calling them failed
		s.clients = append(s.clients[:val-i], s.clients[(val-i)+1:]...)
	}

	return responses[0]
}

func FormatOutcome(outcome *DAS.Outcome) string {
	r := fmt.Sprintf("%v", outcome)
	if len(r) == 0 {
		r = "There is no active auction"
	} else {
		if outcome.Left > 0 {
			if outcome.Bidder != 0 {
				r = fmt.Sprintf("| Auction for '%s' has %vms left\n| Highest bid (by id %v) is %v", outcome.Item, outcome.Left, outcome.Bidder, outcome.Amount)
			} else {
				r = fmt.Sprintf("| Auction for '%s' has %vms left\n| Starting bid is %v", outcome.Item, outcome.Left, outcome.Amount)
			}
		} else {
			if outcome.Bidder != 0 {
				r = fmt.Sprintf("| Auction for '%s' was won (by id %v) for %v", outcome.Item, outcome.Bidder, outcome.Amount)
			} else {
				r = fmt.Sprintf("| Auction for '%s' did not sell\n| Starting bid was %v", outcome.Item, outcome.Amount)
			}
		}
	}
	return r
}

func (s *ReplicaServers) GetResults() *DAS.Outcome {
	query := &DAS.Empty{}

	var responses []*DAS.Outcome
	var remove []int
	if VERBOSE {
		fmt.Println("--- All responses ---")
	}
	for i, r := range s.clients {
		outcome, err := r.Result(s.ctx, query)
		if err != nil {
			remove = append(remove, i)
			if VERBOSE {
				log.Printf("Port %v | %s\n", clientToPort[r], err)
			}
			continue
		}
		if VERBOSE {
			log.Printf("Port %v | %s\n", clientToPort[r], outcome)
		}
		responses = append(responses, outcome)
	}
	if VERBOSE {
		fmt.Println("---------------------")
	}

	for i, val := range remove {
		// remove from ReplicaServers clients these indexes - since calling them failed
		s.clients = append(s.clients[:val-i], s.clients[(val-i)+1:]...)
	}

	return responses[0]
}

func (s *ReplicaServers) StartAuction(start uint64, duration uint32, name string) *DAS.Ack {
	query := &DAS.Item{
		Name:  name,
		Start: start,
		Alive: duration,
	}

	var responses []*DAS.Ack
	var remove []int
	if VERBOSE {
		log.Println("--- All responses ---")
	}
	for i, r := range s.clients {
		ack, err := r.StartAuction(s.ctx, query)
		if err != nil {
			remove = append(remove, i)
			if VERBOSE {
				log.Printf("Port %v | %s\n", clientToPort[r], err)
			}
			continue
		}
		if VERBOSE {
			log.Printf("Port %v | %s\n", clientToPort[r], ack)
		}
		responses = append(responses, ack)
	}
	if VERBOSE {
		fmt.Println("---------------------")
	}

	for i, val := range remove {
		// remove from ReplicaServers clients these indexes - since calling them failed
		s.clients = append(s.clients[:val-i], s.clients[(val-i)+1:]...)
	}

	return responses[0]
}

// sets the logger to use a log.txt file instead of the console
func setLog(id uint32) *os.File {
	filename := fmt.Sprintf("client-%v.txt", id)
	// Clears the log.txt file when a new client is started
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
	log.SetOutput(mw)
	return f
}
