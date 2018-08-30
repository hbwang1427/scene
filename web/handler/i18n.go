package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/chai2010/jsonmap"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

var (
	i18nBundle = &i18n.Bundle{DefaultLanguage: language.English}

	localesMap = map[string]language.Tag{
		"en":         language.English,
		"zh":         language.SimplifiedChinese,
		"zh-hans-cn": language.SimplifiedChinese,
		"zh-hans":    language.SimplifiedChinese,
		"zh-cn":      language.SimplifiedChinese,
		"zh-hant":    language.TraditionalChinese,
		"zh-tw":      language.TraditionalChinese,
		"ja":         language.Japanese,
		"fr":         language.French,
		"ru":         language.Russian,
		"es":         language.Spanish,
		"ko":         language.Korean,
		"ar":         language.Arabic,
	}
)

func loadI18n() {
	dirs, err := ioutil.ReadDir("assets/locales")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("load locales error")
		return
	}

	for _, fi := range dirs {
		if fi.Name() == "." || fi.Name() == ".." {
			continue
		}
		c, err := ioutil.ReadFile(path.Join("assets/locales", fi.Name()))
		if err != nil {
			log.WithFields(log.Fields{"file": fi.Name(), "error": err}).Warn("load locales error")
			continue
		}
		var messages map[string]interface{}
		if err = json.Unmarshal(c, &messages); err != nil {
			log.WithFields(log.Fields{"file": fi.Name(), "error": err}).Warn("load locales error")
			continue
		}
		tag := strings.ToLower(path.Base(fi.Name()))
		if i := strings.Index(tag, "."); i > 0 {
			tag = tag[:i]
		}
		langTag, ok := localesMap[tag]
		if !ok {
			log.WithFields(log.Fields{"lang": tag}).Warn("language tag not found")
			continue
		}
		m := jsonmap.NewJsonMapFromKV(messages, ".").ToFlatMap(".")
		var localeMsgs []*i18n.Message
		for k, v := range m {
			log.WithFields(log.Fields{"ID": k[1:], "msg": v.(string)}).Debug("add locale message")
			localeMsgs = append(localeMsgs, &i18n.Message{
				ID:    k[1:],
				Other: v.(string),
			})
		}
		//log.Printf("add messages:%v %#v\r\n", tag, langTag)
		err = i18nBundle.AddMessages(langTag, localeMsgs...)
		if err != nil {
			log.WithFields(log.Fields{"file": fi.Name(), "error": err}).Warn("add messages error")
		}

	}
}

func init() {
	//log.SetLevel(log.DebugLevel)
	loadI18n()
}

func Tr(args ...interface{}) string {
	if len(args) == 0 {
		return fmt.Sprintf("TrError: no args")
	}
	var language, msgId string
	if len(args) < 2 {
		language = "en"
		msgId = args[0].(string)
		//return fmt.Sprintf("TrError:%#v", args)
	} else {
		var ok bool
		language, ok = args[0].(string)
		if !ok {
			language = "en"
		}
		msgId, ok = args[1].(string)
		if !ok {
			return fmt.Sprintf("invalid message id:%v", args[1])
		}
	}

	localizer := i18n.NewLocalizer(i18nBundle, language)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: msgId})
	if err != nil {
		log.WithFields(log.Fields{
			"msgId":    msgId,
			"language": language,
			"err":      err,
		}).Warn("localize error")
		//try en
		localizer = i18n.NewLocalizer(i18nBundle, "en")
		msg, err = localizer.Localize(&i18n.LocalizeConfig{MessageID: msgId})
		if err != nil {
			log.WithFields(log.Fields{
				"msgId": msgId,
				"err":   err,
			}).Warn("localize error")
			return msgId
		}
	}
	if len(args) >= 2 {
		return fmt.Sprintf(msg, args[2:]...)
	}
	return msg
}

func GinTr(c *gin.Context, args ...interface{}) string {
	userLang := c.GetString("lang")
	if len(userLang) == 0 {
		userLang = "en"
	}
	msgId, ok := args[0].(string)
	if !ok {
		return fmt.Sprintf("TrError:%#v", args)
	}

	localizer := i18n.NewLocalizer(i18nBundle, userLang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: msgId})
	if err != nil {
		log.WithFields(log.Fields{
			"msgId":    msgId,
			"language": userLang,
			"err":      err,
		}).Warn("localize error")
		//try en
		localizer = i18n.NewLocalizer(i18nBundle, "en")
		msg, err = localizer.Localize(&i18n.LocalizeConfig{MessageID: msgId})
		if err != nil {
			log.WithFields(log.Fields{
				"msgId": msgId,
				"err":   err,
			}).Warn("localize error")
			return msgId
		}
	}
	if len(args) >= 1 {
		return fmt.Sprintf(msg, args[1:]...)
	}
	return msg
}
