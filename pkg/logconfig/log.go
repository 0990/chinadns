package logconfig

import (
	"fmt"
	"github.com/0990/chinadns/pkg/util"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
	"sort"
)

var SortKeys = []string{logrus.FieldKeyTime, logrus.FieldKeyLevel, "id", logrus.FieldKeyMsg, "q", "rtt"}

var SortingFunc = func(keys []string) {
	sort.SliceStable(keys, func(i, j int) bool {
		//按照sortKeys的顺序排序keys,sortKeys中索引越小的越靠前,如果不在sortKeys中，则比较字符串大小
		iIndex := util.IndexOfString(SortKeys, keys[i])
		jIndex := util.IndexOfString(SortKeys, keys[j])
		if iIndex != -1 && jIndex != -1 {
			return iIndex < jIndex
		}

		if iIndex != -1 {
			return true
		}

		if jIndex != -1 {
			return false
		}

		return keys[i] < keys[j] //按照字符串大小排序
	})
}

func InitLogrus(name string, maxMB int, level logrus.Level) {

	formatter := &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05",
		SortingFunc:      SortingFunc,
	}

	logrus.SetFormatter(formatter)

	logrus.SetLevel(level)
	logrus.AddHook(NewDefaultHook(name, maxMB))
}

type DefaultHook struct {
	writer io.Writer
	fmt    logrus.Formatter
}

func NewDefaultHook(name string, maxSize int) *DefaultHook {
	formatter := &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05",
		SortingFunc:      SortingFunc,
	}

	writer := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s.log", name),
		MaxSize:    maxSize,
		MaxAge:     7,
		MaxBackups: 7,
		LocalTime:  true,
		Compress:   false,
	}

	return &DefaultHook{
		writer: writer,
		fmt:    formatter,
	}
}

func (p *DefaultHook) Fire(entry *logrus.Entry) error {
	data, err := p.fmt.Format(entry)
	if err != nil {
		return err
	}
	_, err = p.writer.Write(data)
	return err
}

func (p *DefaultHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
