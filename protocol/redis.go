package protocol

import (
	"bufio"
	"duoduo.com/redis-rate/protocol/redis"
	"duoduo.com/redis-rate/protocol/redis/proto"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/secmask/go-redisproto"
	"io"
	"regexp"
	"strings"
)

// RedisStreamFactory implements tcpassembly.StreamFactory
type RedisStreamFactory struct {
	Filter string
	Port   string
	Iface  string
	KeyExp string
	
}

func (h *RedisStreamFactory) GetPort() string {
	return h.Port
}
func (h *RedisStreamFactory) WrapperTcp(tcp *layers.TCP) *layers.TCP {
	//tcp.SYN = true
	return tcp
}
func (h *RedisStreamFactory) GetFace() string {
	return h.Iface
}
func (h *RedisStreamFactory) GetFilter() string {
	return h.Filter
}

// redisStream will handle the actual decoding of http requests.
type redisStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	stat           *redis.StateKey
	exp *regexp.Regexp
}

func (h *RedisStreamFactory) Init() {
	go redis.StateKeyOps.RunRate()
}
func (h *RedisStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	
	hstream := &redisStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		stat:      redis.StateKeyOps,
		exp: regexp.MustCompile(h.KeyExp),
	}
	src, _ := transport.Endpoints()
	sprintf := fmt.Sprintf("%v", src)
	if sprintf == h.Port {
		go hstream.runResponse() // Important... we must guarantee that data from the reader stream is read.
	} else {
		go hstream.runRequest() // Important... we must guarantee that data from the reader stream is read.
	}
	//
	
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}
func (h *redisStream) runRequest() {
	for {
		buf := bufio.NewReader(&h.r)
		parser := redisproto.NewParser(buf)
		command, err := parser.ReadCommand()
		if err != nil {
			if err == io.EOF {
				return
			}
			if err == io.ErrUnexpectedEOF {
				return
			}
			fmt.Println("err %s", err)
		} else {
			count := command.ArgCount()
			if count >= 2 {
				cmd := command.Get(0)
				var cmdArgs = ""
				for i := 1; i < count; i++ {
					cmdArgs += strings.Trim(string(command.Get(i)), " \r\n")
				}
				if strings.ToLower(string(cmd)) == "get" {
					queryKey := string(command.Get(1))
					if h.filterKey(queryKey) {
						h.stat.IncQuery()
						h.stat.HitQuery()
						fmt.Printf("hitQuery queryKey %s cout %d ",queryKey,h.stat.GetHitQueryCount())
					}
				}
				fmt.Printf("cmd %s args %s \r\n", cmd, cmdArgs)
			}
			
		}
	}
}
func (h *redisStream) filterKey(key string) bool {
	return h.exp.Match([]byte(key))
}
func (h *redisStream) runResponse() {
	for {
		reader := proto.NewReader(&h.r)
		for {
			readLine, err := reader.ReadString()
			if err != nil {
				if err == io.EOF {
					return
				}
				if err == io.ErrUnexpectedEOF {
					return
				}
				fmt.Println("err %s", err)
				break
			} else {
			
				printfContent := fmt.Sprintf("respAll %s ", string(readLine))
				have := strings.Contains(printfContent,"wallItems")
			
				if h.stat.GetHitQueryCount() > 0{
					reset := h.stat.ResetHitQuery()
					fmt.Printf("&& have",have)
					if len(string(readLine)) > 0 && reset && readLine != "OK" {
						h.stat.IncHitCount()
					}
				}
				//if strings.Contains(printfContent,"wallItems") {
				//	h.stat.IncHitCount()
				//}
				fmt.Println(printfContent)
			}
		}
		
	}
	
}
func (h *redisStream) runResponse2() {
	for{
		buf := bufio.NewReader(&h.r)
		for{
			bytesContent, err := buf.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				if err == io.ErrUnexpectedEOF {
					return
				}
				fmt.Println("err %s", err)
				return
			}
			content, err := buf.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				if err == io.ErrUnexpectedEOF {
					return
				}
				fmt.Println("err %s", err)
				return
			}
			fmt.Printf("rsp len %s ,content %s", strings.Trim(string(bytesContent)," \r\n"),content )
			
		}
		
		
		//if true {
		//	len, _, err := buf.ReadLine()
		//	if err != nil {
		//		if err == io.EOF {
		//			return
		//		}
		//		if err == io.ErrUnexpectedEOF {
		//			return
		//		}
		//		fmt.Println("err %s", err)
		//		return
		//	}
		//	var rspLen int = 0
		//	if len[0] == '$' {
		//		s := len[1:]
		//		rspLen,_ = strconv.Atoi(string(s))
		//	}else{
		//
		//		rspLen, _ = strconv.Atoi(string(len))
		//	}
		//
		//	if rspLen > 0 {
		//		rsp, _, _ := buf.ReadLine()
		//		println("rspAll " + string(rsp))
		//	}
		//
		//}
	}
	
}
