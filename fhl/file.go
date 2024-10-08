package fhl

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func (f *FHL) LoadDatasetFile(path string) *FHL {
	f.DatasetFile, f.Error = os.Open(path)
	return f
}

func (f *FHL) LoadPrecalFile(path string) *FHL {
	data, err := os.ReadFile(path)
	if err != nil {
		f.Error = err
		return f
	}
	f.Error = json.Unmarshal(data, f)
	f.ArryToCount()
	return f
}

func (f *FHL) SavePrecal() *FHL {
	logrus.Info("SavePrecal")
	file, err := os.Create(PrecalPath)
	if err != nil {
		f.Error = err
		return f
	}
	defer file.Close()
	f.CountToArry()
	var b []byte
	if b, f.Error = json.Marshal(f); f.Error == nil {
		file.Write(b)
	}
	return f
}

// 纠错数据
func SavePrecalErrCorr(x []ErrCorrRecord) error {
	logrus.Info("SavePrecalErrCorr")
	file, err := os.Create(ErrCorrPath)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	count := 0
	for i, rec := range x {
		if i == 0 || rec != x[i-1] {
			count++
			if err := writeErrCorrRecord(w, rec); err != nil {
				return err
			}
		}
	}

	w.Flush()
	logrus.Infoln("save errcorr len:", count)
	return nil
}

func (f *FHL) LoadPrecalErrCorr(path string) *FHL {
	file, err := os.Open(path)
	if err != nil {
		file.Close()
		f.Error = err
		return f
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		f.Error = err
		return f
	}
	f.ErrCorrNums = stat.Size() / RECORD_W

	f.ErrCorrFile = file
	return f
}

func writeErrCorrRecord(w *bufio.Writer, rec ErrCorrRecord) error {
	buf := [RECORD_W]byte{}
	for i := 0; i < HASH_W; i++ {
		buf[i] = byte(rec.Hash >> (i * 8))
	}
	for i := 0; i < ART_IDX_W; i++ {
		buf[HASH_W+i] = byte(rec.ArticleIdx >> (i * 8))
	}
	for i := 0; i < CON_IDX_W; i++ {
		buf[HASH_W+ART_IDX_W+i] = byte(rec.ContentIdx >> (i * 8))
	}
	_, err := w.Write(buf[:])
	return err
}

func (f *FHL) GetArticle(id int) *Article {
	// 读入篇目
	var datasetFileReader = bufio.NewReaderSize(nil, 512)
	f.DatasetFile.Seek(f.Offset[id], io.SeekStart)
	datasetFileReader.Reset(f.DatasetFile)
	s, err := datasetFileReader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	s = s[:len(s)-1] // 去掉换行符
	article, _ := parseArticle(id, s)
	return article
}
