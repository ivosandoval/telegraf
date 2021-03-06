// +build windows
package ping

import (
	"errors"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Windows ping format ( should support multilanguage ?)
var winPLPingOutput = `
Badanie 8.8.8.8 z 32 bajtami danych:
Odpowiedz z 8.8.8.8: bajtow=32 czas=49ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=46ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=48ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=57ms TTL=43

Statystyka badania ping dla 8.8.8.8:
    Pakiety: Wyslane = 4, Odebrane = 4, Utracone = 0
             (0% straty),
Szacunkowy czas bladzenia pakietww w millisekundach:
    Minimum = 46 ms, Maksimum = 57 ms, Czas sredni = 50 ms
`

// Windows ping format ( should support multilanguage ?)
var winENPingOutput = `
Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=52ms TTL=43
Reply from 8.8.8.8: bytes=32 time=50ms TTL=43
Reply from 8.8.8.8: bytes=32 time=50ms TTL=43
Reply from 8.8.8.8: bytes=32 time=51ms TTL=43

Ping statistics for 8.8.8.8:
    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 50ms, Maximum = 52ms, Average = 50ms
`

func TestHost(t *testing.T) {
	trans, rec, avg, min, max, err := processPingOutput(winPLPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, rec, "4 packets were received")
	assert.Equal(t, 50, avg, "Average 50")
	assert.Equal(t, 46, min, "Min 46")
	assert.Equal(t, 57, max, "max 57")

	trans, rec, avg, min, max, err = processPingOutput(winENPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, rec, "4 packets were received")
	assert.Equal(t, 50, avg, "Average 50")
	assert.Equal(t, 50, min, "Min 50")
	assert.Equal(t, 52, max, "Max 52")
}

func mockHostPinger(timeout float64, args ...string) (string, error) {
	return winENPingOutput, nil
}

// Test that Gather function works on a normal ping
func TestPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com", "www.reddit.com"},
		pingHost: mockHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    4,
		"percent_packet_loss": 0.0,
		"average_response_ms": 50,
		"minimum_response_ms": 50,
		"maximum_response_ms": 52,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	tags = map[string]string{"url": "www.reddit.com"}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

var errorPingOutput = `
Badanie nask.pl [195.187.242.157] z 32 bajtami danych:
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.

Statystyka badania ping dla 195.187.242.157:
    Pakiety: Wysłane = 4, Odebrane = 0, Utracone = 4
             (100% straty),
`

func mockErrorHostPinger(timeout float64, args ...string) (string, error) {
	return errorPingOutput, errors.New("No packets received")
}

// Test that Gather works on a ping with no transmitted packets, even though the
// command returns an error
func TestBadPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockErrorHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.amazon.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    0,
		"percent_packet_loss": 100.0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

var lossyPingOutput = `
Badanie thecodinglove.com [66.6.44.4] z 9800 bajtami danych:
Upłynął limit czasu żądania.
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=118ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Upłynął limit czasu żądania.
Odpowiedź z 66.6.44.4: bajtów=9800 czas=119ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=116ms TTL=48

Statystyka badania ping dla 66.6.44.4:
    Pakiety: Wysłane = 9, Odebrane = 7, Utracone = 2
             (22% straty),
Szacunkowy czas błądzenia pakietów w millisekundach:
    Minimum = 114 ms, Maksimum = 119 ms, Czas średni = 115 ms
`

func mockLossyHostPinger(timeout float64, args ...string) (string, error) {
	return lossyPingOutput, nil
}

// Test that Gather works on a ping with lossy packets
func TestLossyPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com"},
		pingHost: mockLossyHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 9,
		"packets_received":    7,
		"percent_packet_loss": 22.22222222222222,
		"average_response_ms": 115,
		"minimum_response_ms": 114,
		"maximum_response_ms": 119,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

// Fatal ping output (invalid argument)
var fatalPingOutput = `
Bad option -d.


Usage: ping [-t] [-a] [-n count] [-l size] [-f] [-i TTL] [-v TOS]
            [-r count] [-s count] [[-j host-list] | [-k host-list]]
            [-w timeout] [-R] [-S srcaddr] [-4] [-6] target_name

Options:
    -t             Ping the specified host until stopped.
                   To see statistics and continue - type Control-Break;
                   To stop - type Control-C.
    -a             Resolve addresses to hostnames.
    -n count       Number of echo requests to send.
    -l size        Send buffer size.
    -f             Set Don't Fragment flag in packet (IPv4-only).
    -i TTL         Time To Live.
    -v TOS         Type Of Service (IPv4-only. This setting has been deprecated
                   and has no effect on the type of service field in the IP Header).
    -r count       Record route for count hops (IPv4-only).
    -s count       Timestamp for count hops (IPv4-only).
    -j host-list   Loose source route along host-list (IPv4-only).
    -k host-list   Strict source route along host-list (IPv4-only).
    -w timeout     Timeout in milliseconds to wait for each reply.
    -R             Use routing header to test reverse route also (IPv6-only).
    -S srcaddr     Source address to use.
    -4             Force using IPv4.
    -6             Force using IPv6.

`

func mockFatalHostPinger(timeout float64, args ...string) (string, error) {
	return fatalPingOutput, errors.New("So very bad")
}

// Test that a fatal ping command does not gather any statistics.
func TestFatalPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockFatalHostPinger,
	}

	p.Gather(&acc)
	assert.False(t, acc.HasMeasurement("packets_transmitted"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("packets_received"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("percent_packet_loss"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("average_response_ms"),
		"Fatal ping should not have packet measurements")
}
