package message

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestCode_String(t *testing.T) {
	for i := Code(0); i < Code(0xff); i++ {
		t.Log(i, i.Category(), i.Details(), int(i))
	}
	t.Log(Code(0xff))
}

func TestMessage_Decode(t *testing.T) {

	for ixi, tst := range []struct {
		hexStream string
		desc      string
		verifier  func(m *Message) (err error)
	}{
		{
			hexStream: "480109d6fc911b24d83d851637636f61702e6d65847061746860",
		},
		{
			hexStream: "480100664d65822107fcfd523d0a63616c69666f726e69756d2e65636c697073652e6f72678b2e77656c6c2d6b6e6f776e04636f7265",
			desc:      "1\t0.000000\t192.168.0.7\t104.196.15.150\tCoAP\t96\tCON, MID:102, GET, TKN:4d 65 82 21 07 fc fd 52, coap://californium.eclipse.org/.well-known/core",
		},
		{
			hexStream: "684500664d65822107fcfd52c128b10a520850ff3c2f6f62733e3b63743d313b6f62733b72743d226f627365727665223b7469746c653d224f627365727661626c65207265736f75726365207768696368206368",
			desc:      "2\t0.407514\t104.196.15.150\t192.168.0.7\tCoAP\t126\tACK, MID:102, 2.05 Content, TKN:4d 65 82 21 07 fc fd 52, Block #0, coap://californium.eclipse.org/.well-known/core",
		},
		{
			hexStream: "480100674d65822107fcfd523d0a63616c69666f726e69756d2e65636c697073652e6f72678b2e77656c6c2d6b6e6f776e04636f7265c112",
			desc:      "3\t0.408799\t192.168.0.7\t104.196.15.150\tCoAP\t98\tCON, MID:103, GET, TKN:4d 65 82 21 07 fc fd 52, Block #1, coap://californium.eclipse.org/.well-known/core",
		},
		{
			hexStream: "684500674d65822107fcfd52c128b11aff616e6765732065766572792035207365636f6e6473222c3c2f6f62732d70756d70696e673e3b6f62733b72743d226f627365727665223b7469746c653d224f62",
			desc:      "4\t0.820249\t104.196.15.150\t192.168.0.7\tCoAP\t123\tACK, MID:103, 2.05 Content, TKN:4d 65 82 21 07 fc fd 52, Block #1, coap://californium.eclipse.org/.well-known/core",
		},
		{
			hexStream: "480100684d65822107fcfd523d0a63616c69666f726e69756d2e65636c697073652e6f72678b2e77656c6c2d6b6e6f776e04636f7265c122",
		},
		{
			hexStream: "684500774d65822107fcfd52c128b2011aff3c2f6c696e6b313e3b69663d22496631223b72743d225479706531205479706532223b7469746c653d224c696e6b2074657374207265736f75726365222c3c2f",
			desc:      "222\t471.449561\t104.196.15.150\t192.168.0.7\tCoAP\t124\tACK, MID:119, 2.05 Content, TKN:4d 65 82 21 07 fc fd 52, Block #17, coap://californium.eclipse.org/.well-known/core",
		},
		{
			hexStream: "684500874d65822107fcfd52c128b20212ff686f64222c3c2f73687574646f776e3e",
			desc:      "305\t880.338273\t104.196.15.150\t192.168.0.7\tCoAP\t76\tACK, MID:135, 2.05 Content, TKN:4d 65 82 21 07 fc fd 52, End of Block #33, coap://californium.eclipse.org/.well-known/core",
		},
		{
			hexStream: "480100884c22b02936d4ff9b3d0a63616c69666f726e69756d2e65636c697073652e6f726730536f6273600101",
			desc:      "306\t880.339826\t192.168.0.7\t104.196.15.150\tCoAP\t87\tCON, MID:136, GET, TKN:4c 22 b0 29 36 d4 ff 9b, coap://californium.eclipse.org/obs",
		},
		{
			hexStream: "684500884c22b02936d4ff9b6307bc7f602105ff31363a33343a3435",
			desc:      "307\t880.648758\t104.196.15.150\t192.168.0.7\tCoAP\t70\tACK, MID:136, 2.05 Content, TKN:4c 22 b0 29 36 d4 ff 9b, coap://californium.eclipse.org/obs (text/plain)",
		},
		{
			hexStream: "4845ddab4c22b02936d4ff9b6307bc82602105ff31363a33353a3030",
			desc:      "311\t898.087553\t104.196.15.150\t192.168.0.7\tCoAP\t70\tCON, MID:56747, 2.05 Content, TKN:4c 22 b0 29 36 d4 ff 9b, coap://californium.eclipse.org/obs (text/plain)",
		},
		{
			hexStream: "6800dda978629a0f5f3f164f",
			desc:      "312\t905.172597\t192.168.0.7\t104.196.15.150\tCoAP\t54\tACK, MID:56745, Empty Message, TKN:78 62 9a 0f 5f 3f 16 4f",
		},
		{
			hexStream: "480100934c22b02936d4ff9b3d0a63616c69666f726e69756d2e65636c697073652e6f72673101536f627360",
			desc:      "471\t2676.857576\t192.168.0.7\t104.196.15.150\tCoAP\t86\tCON, MID:147, GET, TKN:4c 22 b0 29 36 d4 ff 9b, coap://californium.eclipse.org/obs",
		},
		{
			hexStream: "684500934c22b02936d4ff9bc02105ff31373a30343a3430",
			desc:      "472\t2677.169759\t104.196.15.150\t192.168.0.7\tCoAP\t66\tACK, MID:147, 2.05 Content, TKN:4c 22 b0 29 36 d4 ff 9b, coap://californium.eclipse.org/obs (text/plain)",
		},
		{
			hexStream: "480109d6fc911b24d83d851637636f61702e6d65847061746860",
		},
	} {
		var m = new(Message)
		data, err := hex.DecodeString(tst.hexStream)
		desc := strings.Split(tst.desc, "\t")
		t.Logf("[tst#%-3d] decoding  : %v - %v", ixi, desc[0], desc[len(desc)-1])
		if err = m.Decode(data); err != nil {
			t.Fatalf("[tst#%-3d] decoded failed: %v", ixi, err)
		}

		if tst.verifier != nil {
			if err = tst.verifier(m); err != nil {
				t.Fatalf("[tst#%-3d] verfier failed: %v", ixi, err)
			}
		}
		t.Logf("[tst#%-3d] decoded OK: %v", ixi, m.String())
	}
}
