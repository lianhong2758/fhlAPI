package fhl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	return f
	// 打开 JSON 文件
	// file, err := os.Open(path)
	// if err != nil {
	// 	f.Error = err
	// 	return f
	// }
	// defer file.Close()

	// // 创建一个新的 JSON 解码器
	// decoder := json.NewDecoder(bufio.NewReader(file))
	// // 循环读取 JSON 数据
	// for {
	// 	// 逐行解析 JSON 数据
	// 	err := decoder.Decode(f)
	// 	if err != nil {
	// 		// 如果已经读取完所有数据，退出循环
	// 		if err.Error() == "EOF" {
	// 			break
	// 		}
	// 		f.Error = err
	// 		return f
	// 	}
	// }
	// return f
}

func (f *FHL) SavePrecal() *FHL {
	fmt.Println("SavePrecal")
	file, err := os.Create("data/2c-precal.json")
	defer file.Close()
	if err != nil {
		f.Error = err
		return f
	}
	var b []byte
	if b, f.Error = json.Marshal(f); f.Error == nil {
		file.Write(b)
	} 
	return f
}

// 纠错数据
func SavePrecalErrCorr(x []ErrCorrRecord) error {
	println("SavePrecalErrCorr")
	file, err := os.Create("data/2c-errcorr.bin")
	defer file.Close()
	if err != nil {
		return err
	}
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
	println("save errcorr len:", count)
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
