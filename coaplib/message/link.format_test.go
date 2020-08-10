package message_test

import (
	"github.com/hedzr/go-socketlib/coaplib/message"
	"testing"
)

func TestNewLinkFormatParser(t *testing.T) {
	lfp := message.NewLinkFormatParser()
	res, err := lfp.Parse(strTest1)
	if err != nil {
		t.Fatal(err)
	} else if len(res) != 28 {
		t.Fatal("expect 28 elements in res")
	} else {
		t.Logf("res: %v", res)
	}
}

func TestNewLinkFormat(t *testing.T) {
	lf := message.NewLinkFormat()
	err := lf.Parse(strTest2)
	if err != nil {
		t.Fatal(err)
	} else if len(lf.ResArray) != 28 && len(lf.ResArray) != 34 {
		//goland:noinspection ALL
		t.Fatalf("expect 28 elements in res.\n%v", lf)
	} else {
		t.Logf("res: %v", lf)
	}
}

const (
	strTest1 = `</test>;rt="test";ct=0,</validate>;rt="validate";ct=0,</hello>;rt="Type1";ct=0;if="If1",</bl%C3%A5b%C3%A6rsyltet%C3%B8y>;rt="blåbærsyltetøy";ct=0,</sink>;rt="sink";ct=0,</separate>;rt="separate";ct=0,</large>;rt="Type1 Type2";ct=0;sz=1700;if="If2",</secret>;rt="secret";ct=0,</broken>;rt="Type2 Type1";ct=0;if="If2 If1",</weird33>;rt="weird33";ct=0,</weird44>;rt="weird44";ct=0,</weird55>;rt="weird55";ct=0,</weird333>;rt="weird333";ct=0,</weird3333>;rt="weird3333";ct=0,</weird33333>;rt="weird33333";ct=0,</123412341234123412341234>;rt="123412341234123412341234";ct=0,</location-query>;rt="location-query";ct=0,</create1>;rt="create1";ct=0,</large-update>;rt="large-update";ct=0,</large-create>;rt="large-create";ct=0,</query>;rt="query";ct=0,</seg1>;rt="seg1";ct=40,</path>;rt="path";ct=40,</location1>;rt="location1";ct=40,</multi-format>;rt="multi-format";ct=0,</3>;rt="3";ct=50,</4>;rt="4";ct=50,</5>;rt="5";ct=50`
	strTest2 = `</obs>;ct=1;obs;rt="observe";title="Observable resource which changes every 5 seconds",</obs-pumping>;obs;rt="observe";title="Observable resource which changes every 5 seconds",</separate>;title="Resource which cannot be served immediately and which cannot be acknowledged in a piggy-backed way",</large-create>;rt="block";title="Large resource that can be created using POST method",</large-create/4>;ct=0;sz=20,</large-create/5>;ct=0;sz=3,</large-create/6>;ct=0;sz=0,</large-create/7>;ct=0;sz=0,</seg1>;title="Long path resource",</seg1/seg2>;title="Long path resource",</seg1/seg2/seg3>;title="Long path resource",</large-separate>;rt="block";sz=1280;title="Large resource",</obs-reset>,</.well-known/core>,</multi-format>;ct="0 41 50 60";title="Resource that exists in different content formats (text/plain utf8 and application/xml)",</path>;ct=40;title="Hierarchical link description entry",</path/sub1>;title="Hierarchical link description sub-resource",</path/sub2>;title="Hierarchical link description sub-resource",</path/sub3>;title="Hierarchical link description sub-resource",</link1>;if="If1";rt="Type1 Type2";title="Link test resource",</link3>;if="foo";rt="Type1 Type3";title="Link test resource",</link2>;if="If2";rt="Type2 Type3";title="Link test resource",</obs-large>;obs;rt="observe";title="Observable resource which changes every 5 seconds",</validate>;ct=0;sz=35;title="Resource which varies",</test>;title="Default test resource",</large>;rt="block";sz=1280;title="Large resource",</obs-pumping-non>;obs;rt="observe";title="Observable resource which changes every 5 seconds",</query>;title="Resource accepting query parameters",</large-post>;rt="block";title="Handle POST with two-way blockwise transfer",</location-query>;title="Perform POST transaction with responses containing several Location-Query options (CON mode)",</obs-non>;obs;rt="observe";title="Observable resource which changes every 5 seconds",</create1>;ct=1;sz=2;title="Resource which does not exist yet (to perform atomic PUT)",</large-update>;ct=0;rt="block";sz=0;title="Large resource that can be updated using PUT method",</shutdown>`
)
