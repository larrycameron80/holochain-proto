package holochain

import (
	"context"
	"fmt"
	ic "github.com/libp2p/go-libp2p-crypto"
	net "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"strings"
	"testing"
)

func TestNewNode(t *testing.T) {

	node, err := makeNode(1234, "")
	Convey("It should create a node", t, func() {
		So(err, ShouldBeNil)
		So(node.NetAddr.String(), ShouldEqual, "/ip4/127.0.0.1/tcp/1234")
		So(node.HashAddr.Pretty(), ShouldEqual, "QmNN6oDiV4GsfKDXPVmGLdBLLXCM28Jnm7pz7WD63aiwSG")
	})

	Convey("It should send between nodes", t, func() {
		node2, err := makeNode(4321, "node2")
		So(err, ShouldBeNil)
		node.Host.Peerstore().AddAddr(node2.HashAddr, node2.NetAddr, pstore.PermanentAddrTTL)
		var payload string
		node2.Host.SetStreamHandler("/testprotocol/1.0.0", func(s net.Stream) {
			defer s.Close()

			buf := make([]byte, 1024)
			n, err := s.Read(buf)
			if err != nil {
				payload = err.Error()
			} else {
				payload = string(buf[:n])
			}

			_, err = s.Write([]byte("I got: " + payload))

			if err != nil {
				panic(err)
			}
		})

		s, err := node.Host.NewStream(context.Background(), node2.HashAddr, "/testprotocol/1.0.0")
		So(err, ShouldBeNil)
		_, err = s.Write([]byte("greetings"))
		So(err, ShouldBeNil)

		out, err := ioutil.ReadAll(s)
		So(err, ShouldBeNil)
		So(payload, ShouldEqual, "greetings")
		So(string(out), ShouldEqual, "I got: greetings")

	})

}

func makeNode(port int, id string) (*Node, error) {
	listenaddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)
	// use a constant reader so the key will be the same each time for the test...
	r := strings.NewReader(id + "1234567890123456789012345678901234567890")
	key, _, err := ic.GenerateEd25519Key(r)
	if err != nil {
		panic(err)
	}
	return NewNode(listenaddr, key)
}
