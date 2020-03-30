package protocol

import (
	"bufio"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// HttpStreamFactory implements tcpassembly.StreamFactory
type HttpStreamFactory struct{
	Filter string
	Port string
	Iface string
	KeyExp string
}


// httpStream will handle the actual decoding of http requests.
type httpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	exp *regexp.Regexp
}
func (h *HttpStreamFactory) Init()  {
}
func (h *HttpStreamFactory) GetFilter() string{
	return h.Filter
}
func (h *HttpStreamFactory) GetFace() string{
	return h.Iface
}
func (h *HttpStreamFactory) GetPort() string{
	return h.Port
}
func (h *HttpStreamFactory) WrapperTcp(tcp *layers.TCP) *layers.TCP{
	return tcp
}
func (h *HttpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		exp: regexp.MustCompile(h.KeyExp),
	}
	src,_ := transport.Endpoints()
	sprintf := fmt.Sprintf("%v", src)
	if sprintf == h.Port{
		go hstream.runResponse() // Important... we must guarantee that data from the reader stream is read.
	}else{
		go hstream.run() // Important... we must guarantee that data from the reader stream is read.
	}
	
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}
func (h *httpStream) runResponse()  {
	defer func() {
		if v := recover();v == 11 {
			fmt.Printf("resp v: %#v\n",v)
		}
		
	}()
	for{
		buf := bufio.NewReader(&h.r)
		rsp, err := http.ReadResponse(buf,nil)
		if err == io.EOF || err == io.ErrUnexpectedEOF{
			return
		}else if err != nil {
			println("err %s",err.Error())
			return
			
		} else{
			//bodyBytes := tcpreader.DiscardBytesToEOF(rsp.Body)
			//printResponse(rsp,h,bodyBytes)
			bodyBytes, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			printHeader(rsp.Header)
			println(bodyString)
			
			defer rsp.Body.Close()
			
			
		}
	}
	

	
}
func (h *httpStream) run()  {
	for{
		buf := bufio.NewReader(&h.r)
		
		request, e := http.ReadRequest(buf)
	 
		if e != nil {
			if e == io.EOF {
				return
			}
			println("read Req error %s",e.Error())
		}else{
			println("req url %s",request.RequestURI)
		}
		
	}
	
}
func printResponse(resp *http.Response,h *httpStream,bodyBytes int){
	
	fmt.Println("\n\r")
	fmt.Println(resp.Proto, resp.Status)
	printHeader(resp.Header)
 
}
func printHeader(h http.Header){
	for k,v := range h{
		fmt.Println(k,v)
	}
}