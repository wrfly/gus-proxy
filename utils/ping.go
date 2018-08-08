package utils

// Ping host and returns the average rtt
func Ping(ip string) float32 {
	return 10

	// TODO: different operating system has different ping command,
	// can not use go-ping https://github.com/sparrc/go-ping
	// because it has to set extra system settings
	// (sudo sysctl -w net.ipv4.ping_group_range="0   2147483647")

	// alpine:
	// BusyBox v1.28.4 (2018-05-30 10:45:57 UTC) multi-call binary.
	// Usage: ping [OPTIONS] HOST

	// Send ICMP ECHO_REQUEST packets to network hosts

	// 	-4,-6		Force IP or IPv6 name resolution
	// 	-c CNT		Send only CNT pings
	// 	-s SIZE		Send SIZE data bytes in packets (default 56)
	// 	-t TTL		Set TTL
	// 	-I IFACE/IP	Source interface or IP address
	// 	-W SEC		Seconds to wait for the first response (default 10)
	// 			(after all -c CNT packets are sent)
	// 	-w SEC		Seconds until ping exits (default:infinite)
	// 			(can exit earlier with -c CNT)
	// 	-q		Quiet, only display output at start
	// 			and when finished
	// 	-p HEXBYTE	Pattern to use for payload

	// darwin:
	// ping: option requires an argument -- h
	// usage: ping [-AaDdfnoQqRrv] [-c count] [-G sweepmaxsize]
	// 			[-g sweepminsize] [-h sweepincrsize] [-i wait]
	// 			[-l preload] [-M mask | time] [-m ttl] [-p pattern]
	// 			[-S src_addr] [-s packetsize] [-t timeout][-W waittime]
	// 			[-z tos] host
	// 	   ping [-AaDdfLnoQqRrv] [-c count] [-I iface] [-i wait]
	// 			[-l preload] [-M mask | time] [-m ttl] [-p pattern] [-S src_addr]
	// 			[-s packetsize] [-T ttl] [-t timeout] [-W waittime]
	// 			[-z tos] mcast-group
	// Apple specific options (to be specified before mcast-group or host like all options)
	// 			-b boundif           # bind the socket to the interface
	// 			-k traffic_class     # set traffic class socket option
	// 			-K net_service_type  # set traffic class socket options
	// 			-apple-connect       # call connect(2) in the socket
	// 			-apple-time          # display current time

	// cmd := exec.Command("ping", "-A", "-c", "3", "-w", "2", ip)
	// b, err := cmd.Output()
	// if err != nil {
	// 	return 9999
	// }

	// o := string(b)
	// l := strings.LastIndex(o, "=")
	// o = string(o[l+2:])
	// s := strings.Split(o, "/")

	// avg := s[1]

	// f, _ := strconv.ParseFloat(avg, 32)

	// return float32(f)
}
