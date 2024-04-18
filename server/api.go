package server

import (
	"encoding/json"
	"fhlApi/fhl"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/gin-gonic/gin"
)

type RespCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Topic struct {
	ModType string `json:"modtype"`
	Size    int    `json:"size"`
	ID      string `json:"id"`
}
type TopicResp struct {
	ModType       string `json:"modtype"`
	Size          int    `json:"size"`
	ID            string `json:"id"`
	SubjectString string `json:"subjectstring"`
	fhl.Subject   `json:"-"`
}

func GetTopic(f *fhl.FHL) func(*gin.Context) {
	return func(ctx *gin.Context) {
		body, _ := ctx.GetRawData()
		top := new(Topic)
		if err := json.Unmarshal(body, top); err != nil {
			ctx.JSON(200, RespCode{Code: 400, Message: "Bad Body: " + err.Error(), Data: nil})
			return
		}
		fmt.Println("GetTopic: ",top)
		// 检查 size
		var errs string
		switch top.ModType {
		case "A":
			if top.Size != 0 {
				errs = "Incorrect size"
			}
		case "B":
			if top.Size < 5 || top.Size > 9 {
				errs = "Incorrect size"
			}
		case "C":
			if top.Size != 1 && top.Size != 3 {
				errs = "Incorrect size"
			}
		case "D":
			if top.Size < 5 || top.Size > 10 {
				errs = "Incorrect size"
			}
		}
		if errs != "" {
			ctx.JSON(200, RespCode{Code: 400, Message: errs, Data: nil})
			return
		}

		// 生成题目
		var subject fhl.Subject
		switch top.ModType {
		case "A":
			var word string
			if rand.IntN(4) == 0 {
				word = f.GenerateA(0, 1)[0]
			} else {
				word = f.GenerateA(1, 0)[0]
			}
			subject = &fhl.SubjectA{Word: word}
		case "B":
			words := f.GenerateB(top.Size)
			subject = &fhl.SubjectB{Words: []rune(words), CurIndex: 0}
		case "C":
			left, right := f.GenerateC(top.Size, 7+3*top.Size)
			subject = &fhl.SubjectC{
				WordsLeft:  left,
				WordsRight: right,
				UsedRight:  make([]bool, len(right)),
			}
		case "D":
			left, right := f.GenerateD(top.Size)
			subject = &fhl.SubjectD{
				WordsLeft:  left,
				WordsRight: right,
				UsedLeft:   make([]bool, top.Size),
				UsedRight:  make([]bool, top.Size),
			}
		default:
			ctx.JSON(200, RespCode{Code: 400, Message: "ModType ERR : Only Use A|B|C|D", Data: nil})
			return
		}
		t := TopicResp{
			ModType:       top.ModType,
			Size:          top.Size,
			ID:            top.ID,
			SubjectString: subject.Dump(),
			Subject:       subject,
		}
		Games.Set(top.ID, &AnswerResq{TopicResp: t})
		ctx.JSON(200, RespCode{Code: 200, Message: "", Data: &t})
		return
	}
}

type Answer struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type AnswerResq struct {
	TopicResp        //题目
	Text      string `json:"text"` //ans
	//结合mod使用
	// "update": "",  // A: 没有更新
	// "update": "1",   // B: 表示接下来轮到的玩家需要飞的字
	// "update": "4",   // C: 用到的可消除字
	// "update": "0,4", // D: 左边和右边用到的字
	Update      string              `json:"update"`
	HistorySet  []string            `json:"-"`
	History     []fhl.CorrectAnswer `json:"-"`
	HistoryText []string            `json:"history"`

	NextOne fhl.Side `json:"user"`   //下一个该谁回答,默认0开始,也可以根据history%2推算
	Reason  string   `json:"reason"` //游戏消息
}

func UpAnswer(f *fhl.FHL) func(*gin.Context) {
	return func(ctx *gin.Context) {
		body, _ := ctx.GetRawData()
		ans := new(Answer)
		if err := json.Unmarshal(body, ans); err != nil {
			ctx.JSON(200, RespCode{Code: 400, Message: "Bad Body: " + err.Error(), Data: nil})
			return
		}
		fmt.Println("Answer: ", ans)
		//去除逗号
		ans.Text = strings.ReplaceAll(ans.Text, ",", " ")
		ans.Text = strings.ReplaceAll(ans.Text, "，", " ")
		ans.Text = strings.ReplaceAll(ans.Text, ".", "")
		ans.Text = strings.ReplaceAll(ans.Text, "。", "")

		a := Games.Get(ans.ID)
		if a == nil {
			ctx.JSON(200, RespCode{Code: 203, Message: "游戏已结束或已过期...", Data: nil})
			return
		}

		// a.NextOne=a.NextOne^1
		incorrectReason := ""

		texts := strings.Split(ans.Text, " ")
		// 检查长度限制
		totalLen := 0
		for _, s := range texts {
			totalLen += len([]rune(s))
		}
		if totalLen < 4 {
			incorrectReason = "捣浆糊"
		} else if totalLen > 21 {
			incorrectReason = "碎碎念"
		}

		if incorrectReason == "" {
			correct, articleIdx, sentenceIdx := f.LookupText(texts)
			if !correct {
				if articleIdx != -1 {
					incorrectReason = "没背熟"
				} else {
					incorrectReason = "大文豪"
				}
				// println(articleIdx, sentenceIdx)
				_, _ = articleIdx, sentenceIdx
			}
		}

		if incorrectReason == "" {
			for _, s := range texts {
				for _, ss := range a.HistorySet {
					if s == ss {
						incorrectReason = "复读机"
						break
					}
				}
			}
		}

		var kws []fhl.IntPair
		var change interface{}
		if incorrectReason == "" {
			kws, change = a.Subject.Answer(ans.Text, a.NextOne)
			if kws == nil {
				incorrectReason = "不审题"
			}
		}

		if incorrectReason != "" {
			a.Reason = incorrectReason
			ctx.JSON(200, RespCode{Code: 201, Message: "", Data: a})
			Games.Set(a.ID, a)
			return
		}
		//add history
		a.History = append(a.History, fhl.CorrectAnswer{Text: ans.Text, Keywords: kws})
		a.HistorySet = append(a.HistorySet, texts...)

		// 游戏是否已经完成（用完所有的字词）
		if a.Subject.End() {
			// ok
			a.Text = a.History[len(a.History)-1].Dump()
			a.HistoryText = fhl.CorrectAnswerToStringArr(a.History)
			a.Reason = "游戏结束"
			ctx.JSON(200, RespCode{Code: 202, Message: "End", Data: a})
			Games.Delete(a.ID)
		} else {
			//next
			a.NextOne = a.NextOne ^ 1
			a.Update = fmt.Sprint(change)
			a.Text = a.History[len(a.History)-1].Dump()
			a.HistoryText = fhl.CorrectAnswerToStringArr(a.History)
			a.Reason = ""
			ctx.JSON(200, RespCode{Code: 200, Message: "", Data: a})
			Games.Set(a.ID, a)
		}
		return
	}
}
