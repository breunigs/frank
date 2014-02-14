package frank

import (
	"fmt"
	"testing"
)

func TestExtract(t *testing.T) {
	var msgs = make(map[string][]string)
	msgs["Ich finde http://github.com/lol toll, aber http://heise.de besser"] = []string{"http://github.com/lol", "http://heise.de"}
	msgs["dort (http://deinemudda.de) gibts geile pics"] = []string{"http://deinemudda.de"}
	msgs["http://heise.de, letztens gefunden."] = []string{"http://heise.de"}
	msgs["http-rfc ist doof"] = []string{}
	msgs["http://http://foo.de, letztens gefunden."] = []string{"http://http://foo.de"}
	msgs["http://http://foo.de letztens gefunden"] = []string{"http://http://foo.de"}
	msgs["sECuRE: failed Dein Algo nicht auf https://maps.google.de/maps?q=Frankfurt+(Oder)&hl=de ?"] = []string{"https://maps.google.de/maps?q=Frankfurt+(Oder)&hl=de"}

	for from, to := range msgs {
		x := fmt.Sprintf("%v", extract(from))
		to := fmt.Sprintf("%v", to)

		if x != to {
			t.Errorf("extract(%v)\n GOT: %v\nWANT: %v", from, x, to)
		}
	}
}

func TestTitleGet(t *testing.T) {
	var samples = make(map[string]string)
	samples["https://twitter.com/dave_tucker/status/400269131255390210"] = "Dave Tucker (@dave_tucker): This morning the wife asked “Why is your phone issuing you death threats?”. Me: “Oh it’s just my new alarm clock” /cc @CARROT_app"
	samples["http://twitter.com/dave_tucker/status/400269131255390210"] = "Dave Tucker (@dave_tucker): This morning the wife asked “Why is your phone issuing you death threats?”. Me: “Oh it’s just my new alarm clock” /cc @CARROT_app"
	samples["https://twitter.com/Perspective_pic/status/400356645504831489/photo/1"] = "Perspective Pictures (@Perspective_pic): Sorry but this without a doubt the greatest thing ever seen on an air duct https://pbs.twimg.com/media/BY5aP2RIQAAWPl1.jpg:large"
	samples["https://twitter.com/Perspective_pic/status/400356645504831489"] = "Perspective Pictures (@Perspective_pic): Sorry but this without a doubt the greatest thing ever seen on an air duct https://pbs.twimg.com/media/BY5aP2RIQAAWPl1.jpg:large"
	samples["https://twitter.com/quityourjrob/status/405438033853313025/photo/1"] = "Joanna Robinson (@jowrotethis): How to tell if a toy is for boys or girls. https://pbs.twimg.com/media/BaBnvl5CYAAyYzm.jpg:large"
	samples["https://twitter.com/rechelon/status/431242278221275137"] = "William Gillis (@rechelon): @SebastosPublius @jfsmith23 Yep. Godesky had gathered a large following back then and was more sane than Zerzan & less terrible than Jensen."
	samples["https://twitter.com/thejeremyvine/status/433607774375649280"] = "Jeremy Vine (@theJeremyVine): The internet was invented so someone could ask this question - and get an answer: https://pbs.twimg.com/media/BgR7-TQCIAAE4fm.jpg:large"
	samples["http://twitter.com/thejeremyvine/status/433607774375649280"] = "Jeremy Vine (@theJeremyVine): The internet was invented so someone could ask this question - and get an answer: https://pbs.twimg.com/media/BgR7-TQCIAAE4fm.jpg:large"

	for url, title := range samples {
		x, _, _ := TitleGet(url)
		if x != title {
			t.Errorf("TitleGet(%v)\n GOT: ||%v||\nWANT: ||%v||", url, x, title)
		}
	}
}

func TestClean(t *testing.T) {
	if x := clean("x‏‎​   x‏"); x != "x x" {
		t.Errorf("clean does not remove all whitespace/non-printable chars (got: %v)", x)
	}

	if x := clean(" trim "); x != "trim" {
		t.Errorf("clean does not trim properly (got: %v)", x)
	}
}
