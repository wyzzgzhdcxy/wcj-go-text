package utils

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// exist 检查路径（文件或目录）是否存在
func exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CategorizeFiles 根据提供的字符串列表（可能是文件名的某种模式或标签）来归类文件
// 读取指定目录中的文件列表。
// 遍历文件列表，并检查每个文件名是否与提供的字符串列表中的任何一个匹配。
// 根据匹配情况，将文件归类到不同的集合或切片中。
func CategorizeFiles(dir string, categories []string) core.MySet {
	set := core.MySet{}
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("读取目录异常:%s", dir)
		return set
	}

	for _, file := range files {
		filename := file.Name()
		if file.IsDir() {
			continue
		}
		for _, category := range categories {
			// 这里假设我们基于文件名前缀来归类，你可以根据需要修改匹配逻辑
			if strings.HasPrefix(filename, category) {
				dirPath := filepath.Join(dir, category)
				if !exist(dirPath) {
					err := os.MkdirAll(dirPath, os.ModePerm)
					if err != nil {
						log.Printf("创建文件夹异常:%s", dirPath)
					}
				}
				err := os.Rename(filepath.Join(dir, filename), filepath.Join(dirPath, filename))
				//set.Add(category)
				if err != nil {
					log.Printf("重命名文件异常:%s", filepath.Join(dir, filename))
				}
				break // 如果一个文件只属于一个类别，则找到后跳出内层循环
			}
		}
	}
	return set
}

// GetFilePrefixesInDir 获取目录下的文件列表的文件名的前缀，用-隔开
func GetFilePrefixesInDir(dir string, count int) []string {
	if count == 0 {
		count = 20
	}
	m := make(map[string]int)
	// 遍历目录
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 忽略目录和子目录
		if !info.IsDir() {
			// 获取文件名
			filename := filepath.Base(path)
			if strings.ContainsAny(filename, "-") && core.ContainsChinese(filename) {
				// 去掉扩展名获取前缀
				prefix := strings.Split(filename, "-")[0]
				if len(prefix) < 10 {
					prefix = strings.TrimSpace(prefix)
					// 添加前缀到列表中
					if _, exists := m[prefix]; exists {
						m[prefix]++
					} else {
						m[prefix] = 1
					}
				}
			}
		}
		return nil
	})

	// 创建一个切片来存储键值对
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	// 使用sort.Slice和自定义的比较函数对切片进行排序（按年龄升序）
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	// 按照值对切片进行排序
	if len(ss) < count {
		count = len(ss)
	}
	r := make([]string, 0)
	for i := range ss[:count] {
		r = append(r, ss[i].Key)
	}
	return r
}

// 定义一个类型，用于存储键值对，并实现 sort.Interface 接口
type kv struct {
	Key   string
	Value int
}
