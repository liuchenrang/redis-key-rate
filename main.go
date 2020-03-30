package main

import (
    "duoduo.com/redis-rate/protocol"
    "duoduo.com/redis-rate/protocol/redis"
    "flag"
    "fmt"
    
    //"io"
    "log"
    "time"
    
    "github.com/google/gopacket"
    "github.com/google/gopacket/examples/util"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/pcap"
    "github.com/google/gopacket/tcpassembly"
)
var port = "1800"
var ifaceName = "lo0"
var filterRule = "tcp and port " + port
var iface = flag.String("i", ifaceName, "Interface to get packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 1600, "SnapLen for pcap packet capture")
var filter = flag.String("f", filterRule, "BPF filter for pcap")
var parseProtocol = flag.String("p", "redis", "")
var keyExp = flag.String("keyExp", "", "")
var logAllPackets = flag.Bool("v", false, "Logs every packet in great detail")

// Build a simple HTTP request parser using tcpassembly.StreamFactory and tcpassembly.Stream interfaces


func main() {
    defer util.Run()()
    var handle *pcap.Handle
    var err error
    // Set up assembly
    var streamFactory protocol.PtoFace
    if *parseProtocol == "http" {
        streamFactory = &protocol.HttpStreamFactory{"tcp and port 1800","1800",*iface,*keyExp}
    }else{
        streamFactory = &protocol.RedisStreamFactory{"tcp and port 6379","6379",*iface, *keyExp}
    
    }
    streamFactory.Init()
    // Set up pcap packet capture
    if *fname != "" {
        log.Printf("Reading from pcap dump %q", *fname)
        handle, err = pcap.OpenOffline(*fname)
    } else {
        log.Printf("Starting capture on interface %q", *iface)
        handle, err = pcap.OpenLive(*iface, int32(*snaplen), true, pcap.BlockForever)
    }
    if err != nil {
        log.Fatal(err)
    }
    
    if err := handle.SetBPFFilter(streamFactory.GetFilter()); err != nil {
        log.Fatal(err)
    }
    

    streamPool := tcpassembly.NewStreamPool(streamFactory)
    assembler := tcpassembly.NewAssembler(streamPool)
    
    log.Println("reading in packets")
    // Read in packets, pass to assembler.
    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    packets := packetSource.Packets()
    ticker := time.Tick(time.Minute)
    for {
        select {
        case packet := <-packets:
            // A nil packet indicates the end of a pcap file.
            if packet == nil {
                return
            }
            if *logAllPackets {
                log.Println(packet)
            }
            if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
                log.Println("Unusable packet")
                continue
            }
            tcp := packet.TransportLayer().(*layers.TCP)
            tcp = streamFactory.WrapperTcp(tcp)
            assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
            rate := redis.StateKeyOps.CalcRate()
            fmt.Printf("rate %s \r\n",rate)
        case <-ticker:
            // Every minute, flush connections that haven't seen activity in the past 2 minutes.
            assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
        }
    }
    
}