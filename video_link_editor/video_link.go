package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/mgutz/ansi"
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"

	"encoding/json"
	"fmt"
	"os"
)

type OgParam struct {
	OgURL         string `json:"og:url"`
	OgType        string `json:"og:type"`
	OgTitle       string `json:"og:title"`
	OgDescription string `json:"og:description"`
	OgSitename    string `json:"og:site_name"`
	OgImage       string `json:"og:image"`
}

type VideoLink struct {
	URL string  `json:"url"`
	Og  OgParam `json:"og"`
}

type VideoLinkObject struct {
	Stage           string
	EntryCd         string
	Bkt             *s3.Bucket
	BeforeVideoLink VideoLink
	AfterVideoLink  VideoLink
}

const MIME_TYPE = "text/plain"

func NewVideoLinkObject(stageStr, entry_cd string) *VideoLinkObject {
	vlo := &VideoLinkObject{}
	vlo.Stage = stageStr
	vlo.EntryCd = entry_cd
	// add env
	// export AWS_ACCESS_KEY_ID=xxx
	// export AWS_SECRET_ACCESS_KEY=xxx
	auth, err := aws.EnvAuth()
	if err != nil {
		fmt.Println("AWSの環境変数が取得できません：", err)
		return nil
	}
	s3client := s3.New(auth, aws.APNortheast)
	vlo.Bkt = s3client.Bucket(fmt.Sprintf("classi-%s-app", *stage))
	return vlo
}

func (self *VideoLinkObject) Bucket2Struct() error {
	path := fmt.Sprintf("contentsbox/%s", self.EntryCd)
	content, err := self.Bkt.Get(path)
	if err != nil {
		return err
	}
	info := fmt.Sprintf("teaching video information: %s", self.EntryCd)
	fmt.Println(ansi.Color(info, "white+b"))
	fmt.Println(string(content))

	err = json.Unmarshal([]byte(content), &self.BeforeVideoLink)
	if err != nil {
		return err
	}
	return nil
}

func (self *VideoLinkObject) GenerateAfterVideoLink(url string) error {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return err
	}
	ogp := &OgParam{}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		op, _ := s.Attr("property")
		con, _ := s.Attr("content")
		if op == "og:description" {
			ogp.OgDescription = con
		}
		if op == "og:type" {
			ogp.OgType = con
		}
		if op == "og:title" {
			ogp.OgTitle = con
		}
		if op == "og:url" {
			ogp.OgURL = con
		}
		if op == "og:site_name" {
			ogp.OgSitename = con
		}
		if op == "og:image" {
			ogp.OgImage = con
		}
	})

	self.AfterVideoLink = VideoLink{
		url,
		OgParam{
			ogp.OgURL,
			ogp.OgType,
			ogp.OgTitle,
			ogp.OgDescription,
			ogp.OgSitename,
			ogp.OgImage}}
	return nil
}

func (self *VideoLinkObject) ChangeVL() (count int) {
	//fmt.Println(self.BeforeVideoLink.URL, self.AfterVideoLink.URL)
	if len(self.AfterVideoLink.URL) > 0 {
		self.BeforeVideoLink.URL = self.AfterVideoLink.URL
		self.BeforeVideoLink.Og.OgURL = self.AfterVideoLink.URL
		count = count + 1
	}

	if self.AfterVideoLink.URL == "0" {
		self.BeforeVideoLink.URL = ""
		self.BeforeVideoLink.Og.OgURL = ""
		count = count + 1
	}

	if len(self.AfterVideoLink.Og.OgType) > 0 {
		self.BeforeVideoLink.Og.OgType = self.AfterVideoLink.Og.OgType
		count = count + 1
	}

	if self.AfterVideoLink.Og.OgType == "0" {
		self.BeforeVideoLink.Og.OgType = ""
		count = count + 1
	}

	if len(self.AfterVideoLink.Og.OgDescription) > 0 {
		self.BeforeVideoLink.Og.OgDescription = self.AfterVideoLink.Og.OgDescription
		count = count + 1
	}

	if self.AfterVideoLink.Og.OgDescription == "0" {
		self.BeforeVideoLink.Og.OgDescription = ""
		count = count + 1
	}

	if len(self.AfterVideoLink.Og.OgTitle) > 0 {
		self.BeforeVideoLink.Og.OgTitle = self.AfterVideoLink.Og.OgTitle
		count = count + 1
	}

	if self.AfterVideoLink.Og.OgTitle == "0" {
		self.BeforeVideoLink.Og.OgTitle = ""
		count = count + 1
	}

	if len(self.AfterVideoLink.Og.OgSitename) > 0 {
		self.BeforeVideoLink.Og.OgSitename = self.AfterVideoLink.Og.OgSitename
		count = count + 1
	}

	if self.AfterVideoLink.Og.OgSitename == "0" {
		self.BeforeVideoLink.Og.OgSitename = ""
		count = count + 1
	}

	if len(self.AfterVideoLink.Og.OgImage) > 0 {
		self.BeforeVideoLink.Og.OgImage = self.AfterVideoLink.Og.OgImage
		count = count + 1
	}
	if self.AfterVideoLink.Og.OgImage == "0" {
		self.BeforeVideoLink.Og.OgImage = ""
		count = count + 1
	}

	return
}

func (self *VideoLinkObject) Write(dry bool) {
	afterString, err := json.Marshal(self.BeforeVideoLink)
	if err != nil {
		fmt.Println("jsonに変換できません", err)
		os.Exit(1)
	}
	if dry == false {
		filename := fmt.Sprintf("contentsbox/%s", self.EntryCd)
		self.Bkt.Put(filename, afterString, MIME_TYPE, s3.BucketOwnerFull)
		fmt.Println(ansi.Color("以下に変更しました", "green+b"))
	} else {
		fmt.Println(ansi.Color("以下に変更しませんでした", "red+b"))
	}

	fmt.Println(string(afterString), "\n")

}
