package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"sync"
	"regexp"
	"math/rand"

	// ui "github.com/gizak/termui" // maybe later 
	"github.com/codegangsta/cli"
	"github.com/parnurzeal/gorequest"
)

type empty struct{}

func main() {
	tool := cli.NewApp()
	tool.Name = "strain"
	tool.Usage = "Throws a configurable number of GET requests at an HTTP/S server"
	tool.Version = "0.1.0"
	tool.Flags = []cli.Flag {
		cli.IntFlag{
			Name: "workers, w",
			Value: 10,
			Usage: "Number of workers",
		},
		cli.IntFlag{
			Name: "repeat, n",
			Value: -1,
			Usage: "How often each worker repeats before terminating",
		},
		cli.DurationFlag{
			Name: "rate, r",
			Value: 200 * time.Millisecond,
			Usage: "How often each worker pounds the server",
		},
		cli.BoolFlag {
			Name: "quiet, q",
			Usage: "Don't give periodic reports",
		},
		cli.DurationFlag {
			Name: "report_rate",
			Value: 1 * time.Second,
			Usage: "How often to produce periodic reports",
		},
	}
	tool.Action = func(c *cli.Context) {
		rand.Seed(time.Now().UnixNano())

		nWorkers := c.Int("workers")
		repeat   := c.Int("repeat")
		rate     := c.Duration("rate")
		rrate    := c.Duration("report_rate")
		quiet    := c.Bool("quiet")
		
		var url string
		if len(c.Args()) > 0 {
			url = c.Args()[0]
		} else {
			cli.ShowAppHelp(c)
			os.Exit(0)
		}
		var pass, fail, err chan empty
		if !quiet {
			pass = make(chan empty, nWorkers * int(rrate/rate))
			fail = make(chan empty, nWorkers * int(rrate/rate))
			err  = make(chan empty, nWorkers * int(rrate/rate))

			defer func() {
				close(pass)
				close(fail)
				close(err)
			}()
		}

		var wg sync.WaitGroup
		wg.Add(nWorkers)
		for i := nWorkers; i > 0; i-- {
			go launchWorker(
				&wg, url, rate, repeat,
				pass, fail, err,
				quiet,
			)
		}
		mod := time.Duration(int64((0.1 * float64(rate) * 1.5)))
		log.Printf("Launched %d workers; requests centered on %s (σ±%s)\n", nWorkers, rate.String(), mod.String())
		if !quiet {
			go reportState(rrate, pass, fail, err, nWorkers)
		}

		wg.Wait()
	}
	tool.Run(os.Args)
}

func launchWorker(
	wg *sync.WaitGroup, url string, rate time.Duration, repeat int, 
	pass, fail, err chan empty,
	quiet bool,
) {

	defer wg.Done()
	
	mod := time.Duration(int64(rand.NormFloat64() * 0.1 * float64(rate) * 1.5))
	tick := time.Tick(mod + rate)
	for repeat != 0 {
		select {
		case <- tick:
			resp, _, e := gorequest.New().Get(url).End()
			if repeat > 0 { repeat-- }
			if !quiet {
				switch {
				case e != nil:
					if resp != nil {
						err  <- empty{}
						continue
					}
					fallthrough
				case resp.StatusCode >= 400:
					fail <- empty{}
				default:
					pass <- empty{}
				}
			}
		}
	}
}

var dxp = regexp.MustCompile(`^1(ns|us|µs|ms|s|m|h)$`)
func reportState(rrate time.Duration, pass, fail, err chan empty, nWorkers int) {
	tick := time.Tick(rrate)
	num := 0
	strate := dxp.ReplaceAllString(rrate.String(), "$1")
	if strate == "m" {
		strate = "min"
	}
	for {
		select {
		case <- tick:
			ok    := len(pass)
			bad   := len(fail)
			errs  := len(err)
			total := ok+bad+errs
			var faces []string
			switch {
			case bad > errs: faces = kmojiBad
			case bad < errs: faces = kmojiSad
			case total == 0:
				faces = kmojiSad
			default:
				if bad > 0 { faces = kmojiBad
				} else {     faces = kmojiGood }
			}
			face := faces[rand.Intn(len(faces))]
			if rand.Intn(100) > 15 {
				face = ""
			}
			num++
			if num > 99999 { num = 99999 }
			if total == 0 {
				fmt.Printf(
					"[t%05d] %d workers, but no responses received ??? %s\n", 
					num, nWorkers, face,
				)
				continue
			}

			fmt.Printf(
				"[t%05d] %5d requests/%s: %5d pass, %5d fail, %5d errors %s\n", 
				num, total, strate, ok, bad, errs, face,
			)
			// Unwind buffers
			for i := ok; i > 0; i-- {
				<- pass		
			}
			for i := bad; i > 0; i-- {
				<- fail		
			}
			for i := errs; i > 0; i-- {
				<- err		
			}
		}
	}
}

var kmojiGood = []string {
	`o(〃＾▽＾〃)o`, `o(≧∇≦o)`, `(๑˃̵　ᴗ　˂̵)و`, `（˶′◡‵˶）`, "( ´ ▽ ` )b",
	`(￣一*￣)b`, `（￣ー￣）`, "╰(*´︶`*)╯", `٩(๑❛ᴗ❛๑)۶`, `°˖✧◝(⁰▿⁰)◜✧˖°`,
	`(ˆ ڡ ˆ)`, `♫ヽ(゜∇゜ヽ)♪♬(ノ゜∇゜)ノ♩♪`, `♪♪(o*゜∇゜)o～♪♪`, `٩(๑❛ᴗ❛๑)۶`, `＼＼\(۶•̀ᴗ•́)۶//／／`,
	`ᕦ(ò_óˇ)ᕤ`, `ᕙ(＠°▽°＠)ᕗ`, `（￣ー￣）`, `٩(*ゝڡゝ๑)۶`, `( ´･ω･)人(・ω・｀ )`,
	`(｀・ω・´)`, `˭̡̞(◞⁎˃ᆺ˂)◞*✰`, `╰(✧∇✧╰)`, `ヽ༼>ل͜<༽ﾉ`, `(๑>ᴗ<๑)`,
	`(๑•̀ㅂ•́)و`, `(*•̀ᴗ•́*)و ̑̑`, `(ง ͠ ͠° ل͜ °)ง`, `(۶•̀ᴗ•́)۶`, `٩(๑˃̵ᴗ˂̵)و`,
}

var kmojiBad = []string {
	`(๑•̀ㅁ•́๑)✧`, `(•ˋ _ ˊ•)`, `(｀Д´)`, "щ(`Д´щ;)", `(ʘ言ʘ╬)`,
	`( ◉◞౪◟◉)`, `(´⊙◞⊱​◟⊙｀)​`, `( ͡° ͜ʖ ͡°)`, `┬┴┬┴┤(･_├┬┴┬┴`, `(；゜○゜)ア`,
	`(╯°□°）╯︵ ┻━┻`, `(┛◉Д◉)┛彡┻━┻`, `(ノಠ益ಠ)ノ彡┻━┻`, `(ﾉ≧∇≦)ﾉ ﾐ ┸━┸`, `(۶ૈ ᵒ̌ Дᵒ̌)۶ૈ=͟͟͞͞ ⌨`,
	`(｡´・ω・｀｡)Maybe I should send 601 emails next time..`, `ლ(;; ิ益 ิ;‘ლ)`, `(｡☉︵ ಠ╬)`, `ー(￣～￣)ξ`, `（￣＾￣）`,
	`༼ つ ͠° ͟ ͟ʖ ͡° ༽つ`, `ヽ(｀⌒´メ)ノ`, `╰༼=ಠਊಠ=༽╯`, `!! o(*≧д≦)o))`, `┗(•̀へ •́ ╮ )`,
}

var kmojiSad = []string {
	"(´；ω；`)", `o（ｉДｉ）o`, `o(╥﹏╥)`, `(╯︵╰,)`, `（；へ：）`,
	`(´。＿。｀)`, `(ᗒᗩᗕ)`, `(╯︵╰)`, `(◞ ‸ ◟ㆀ)`, `(´∩｀。)`,
	`Σ(°Д°υ)`, `(๑•﹏•)⋆* ⁑⋆*`, `「(°ヘ°)`, "(´◑ω◐`)", `(*´﹃｀*)`,
	`ヘ（。□°）ヘ`, `ヘ(´° □°)ヘ┳━┳`, `What to～(´･ω･｀～)(～´･ω･｀)～～(´･ω･｀～)(～´･ω･｀)～do…`, `(╯⊙ ⊱⊙╰ )`, `┏( .-. ┏ ) ┓`,
	`||Φ|(|´･|ω|･｀|)|Φ||`, `(」゜ロ゜)」`, "(๑ ́ᄇ`๑)", `(。ヘ°)`, `(๑°艸°๑)`,
}

