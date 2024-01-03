# GoUDPflood

This repository contains a Go application designed to create and send UDP packets with a raw socket. It's intended for educational purposes to understand network protocols and raw socket programming in Go.

## Disclaimer

This tool is provided for educational purposes only. Any misuse of this software will not be the responsibility of the author or of any affiliates. The end-user is solely responsible for complying with all applicable laws and regulations in their jurisdiction.

## Prerequisites

- Go 1.20
- Root or Administrator privileges (for raw socket operations)
- Wireshark (for packet analysis, optional)

## Installation

Clone the repository to your local machine:

```sh
git clone https://github.com/araujo88/GoUDPflood.git
cd GoUDPflood
```

## Usage

To run the application, use the following command:

```sh
sudo go run main.go <destination IP> <port number> <data payload>
```

Replace `<destination IP>`, `<port number>`, and `<data payload>` with the appropriate values.

Example:

```sh
sudo go run main.go 127.0.0.1 8080 "Hello, World"
```

## Features

- UDP packet crafting with customizable payload.
- IP and UDP checksum calculation.
- IP header inclusion with the `IP_HDRINCL` option.
- Spoofed source IP address generation.

## How It Works

The application crafts a UDP packet with a manually constructed IP header, followed by a UDP header and a data payload. It sets the `IP_HDRINCL` socket option to signal the kernel that the packet will include the IP header. The checksums for both IP and UDP headers are calculated to ensure the packet's integrity.

## Contributing

Contributions to this project are welcome. Please fork the repository and submit a pull request with your features or fixes.

## License

This project is licensed under the GPL License - see the [LICENSE](LICENSE) file for details.

## Contact

If you have any questions or feedback, please open an issue in the GitHub repository issue tracker.
