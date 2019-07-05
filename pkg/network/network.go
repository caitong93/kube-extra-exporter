package network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

type StatsProvider interface {
	GetStats(rootFs string, pid int) (*Stats, error)
}

func NewStatsProvider() StatsProvider {
	return &defaultProvider{}
}

type defaultProvider struct {
}

func (p *defaultProvider) GetStats(rootFs string, pid int) (*Stats, error) {
	tcpStat, err := tcpStatsFromProc(rootFs, pid, "net/tcp")
	if err != nil {
		return nil, fmt.Errorf("err get tcp stats from pid %v: %v", pid, err)
	}

	tcp6Stat, err := tcpStatsFromProc(rootFs, pid, "net/tcp6")
	if err != nil {
		return nil, fmt.Errorf("err get tcp stats from pid %v: %v", pid, err)
	}

	return &Stats{
		Tcp:  tcpStat,
		Tcp6: tcp6Stat,
	}, nil
}

func tcpStatsFromProc(rootFs string, pid int, file string) (TcpStat, error) {
	tcpStatsFile := path.Join(rootFs, "proc", strconv.Itoa(pid), file)

	tcpStats, err := scanTcpStats(tcpStatsFile)
	if err != nil {
		return tcpStats, fmt.Errorf("couldn't read tcp stats: %v", err)
	}

	return tcpStats, nil
}

func scanTcpStats(tcpStatsFile string) (TcpStat, error) {
	var stats TcpStat

	data, err := ioutil.ReadFile(tcpStatsFile)
	if err != nil {
		return stats, fmt.Errorf("failure opening %s: %v", tcpStatsFile, err)
	}

	tcpStateMap := map[string]uint64{
		"01": 0, //ESTABLISHED
		"02": 0, //SYN_SENT
		"03": 0, //SYN_RECV
		"04": 0, //FIN_WAIT1
		"05": 0, //FIN_WAIT2
		"06": 0, //TIME_WAIT
		"07": 0, //CLOSE
		"08": 0, //CLOSE_WAIT
		"09": 0, //LAST_ACK
		"0A": 0, //LISTEN
		"0B": 0, //CLOSING
	}

	reader := strings.NewReader(string(data))
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)

	// Discard header line
	if b := scanner.Scan(); !b {
		return stats, scanner.Err()
	}

	for scanner.Scan() {
		line := scanner.Text()

		state := strings.Fields(line)
		// TCP state is the 4th field.
		// Format: sl local_address rem_address st tx_queue rx_queue tr tm->when retrnsmt  uid timeout inode
		tcpState := state[3]
		_, ok := tcpStateMap[tcpState]
		if !ok {
			return stats, fmt.Errorf("invalid TCP stats line: %v", line)
		}
		tcpStateMap[tcpState]++
	}

	stats = TcpStat{
		Established: tcpStateMap["01"],
		SynSent:     tcpStateMap["02"],
		SynRecv:     tcpStateMap["03"],
		FinWait1:    tcpStateMap["04"],
		FinWait2:    tcpStateMap["05"],
		TimeWait:    tcpStateMap["06"],
		Close:       tcpStateMap["07"],
		CloseWait:   tcpStateMap["08"],
		LastAck:     tcpStateMap["09"],
		Listen:      tcpStateMap["0A"],
		Closing:     tcpStateMap["0B"],
	}

	return stats, nil
}

type Stats struct {
	Tcp  TcpStat
	Tcp6 TcpStat
}

type TcpStat struct {
	// Count of TCP connections in state "Established"
	Established uint64
	// Count of TCP connections in state "Syn_Sent"
	SynSent uint64
	// Count of TCP connections in state "Syn_Recv"
	SynRecv uint64
	// Count of TCP connections in state "Fin_Wait1"
	FinWait1 uint64
	// Count of TCP connections in state "Fin_Wait2"
	FinWait2 uint64
	// Count of TCP connections in state "Time_Wait
	TimeWait uint64
	// Count of TCP connections in state "Close"
	Close uint64
	// Count of TCP connections in state "Close_Wait"
	CloseWait uint64
	// Count of TCP connections in state "Listen_Ack"
	LastAck uint64
	// Count of TCP connections in state "Listen"
	Listen uint64
	// Count of TCP connections in state "Closing"
	Closing uint64
}
