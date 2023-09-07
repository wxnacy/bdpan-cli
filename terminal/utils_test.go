package terminal

import (
	"testing"
)

func TestOmitString(t *testing.T) {
	var text, origin string
	origin = "wxnacy"
	text = OmitStringRight(origin, 5)
	if text != "wx..." {
		t.Errorf("%s is Error", text)
	}
	text = OmitStringMid(origin, 5)
	if text != "w...y" {
		t.Errorf("%s is Error", text)
	}

	origin = "wxna"
	text = OmitStringRight(origin, 5)
	if text != "wxna" {
		t.Errorf("%s is Error", text)
	}

	origin = "你好s老温"
	text = OmitStringRight(origin, 5)
	if text != "你..." {
		t.Errorf("%s is Error", text)
	}
	text = OmitStringMid(origin, 7)
	if text != "你...温" {
		t.Errorf("%s is Error", text)
	}
	text = OmitString(origin, 7)
	if text != "你...温" {
		t.Errorf("%s is Error", text)
	}

	origin = "你好s"
	text = OmitStringRight(origin, 5)
	if text != "你好s" {
		t.Errorf("%s is Error", text)
	}
	text = OmitStringMid(origin, 5)
	if text != "你好s" {
		t.Errorf("%s is Error", text)
	}
}
