package utils

import (
	"fmt"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var indexFilenameMd5 = "6a992d5529f459a44fee58c733255e86"

// RecoveryFilename /*
func RecoveryFilename(dir string) []string {
	var myMap = make(map[string]string) //md5值和文件名的对应关系
	if b, err := os.ReadFile(dir + "/" + indexFilenameMd5); err == nil {
		arr := strings.Split(string(b), "\r\n")
		for _, fileItem := range arr {
			fileItem = strings.TrimSpace(fileItem)
			// 行格式为 "32位md5-base64文件名"，过短的行（截断/LF结尾等）直接跳过避免越界
			if len(fileItem) < 33 || fileItem[32] != '-' {
				continue
			}
			myMap[fileItem[0:32]] = core.Base64DecodeStr(fileItem[33:])
		}

		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if core.IsDir(path) {
				return nil
			}
			if info.Name() == indexFilenameMd5 {
				return nil
			}
			if core.MapExistKey(&myMap, info.Name()) {
				if err = os.Rename(path, filepath.Dir(path)+"/"+myMap[info.Name()]); err != nil {
					fmt.Println("重命名文件时发生错误:", err)
					return nil
				}
			}
			return nil
		})
		//删除密码文件
		if r, _ := core.PathExists(dir + "/" + indexFilenameMd5); r {
			core.DeleteFile(dir + "/" + indexFilenameMd5)
		}
	}
	return GetFilepathList(dir)
}

func EncryptFilename(dir string) []string {
	if r, _ := core.PathExists(dir + "/" + indexFilenameMd5); r {
		return GetFilepathList(dir)
	} else {
		FileRenameAllDir(dir)
	}
	return GetFilepathList(dir)
}

func FileRenameAllDir(dir string) {
	var builder strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if core.IsDir(path) {
			return nil
		}
		if info.Name() == indexFilenameMd5 {
			return nil
		}
		md5Str := core.Md5(info.Name())
		builder.WriteString(md5Str + "-" + core.Base64EncodeStr(info.Name()) + "\r\n")
		if err = os.Rename(path, filepath.Dir(path)+"/"+md5Str); err != nil {
			fmt.Println("重命名文件时发生错误:", err)
			return nil
		}
		return nil
	})
	if err != nil {
		return
	}
	if err := os.WriteFile(dir+"/"+core.Md5("index"), []byte(builder.String()), 0666); err != nil {
		fmt.Println(err)
	}
}

func GetFilepathList(dir string) []string {
	var pathList []string
	fnList := core.ListFileName(dir, "", false, true)
	for _, fn := range fnList {
		if fn == indexFilenameMd5 {
			continue
		}
		pathList = append(pathList, dir+"/"+fn)
	}
	return pathList
}

// CleanFilename 清理文件名中的特殊字符
// 清理规则：只保留字母、数字、下划线、中文，去掉所有特殊字符（包括空格、引号、问号等）
func CleanFilename(dir string) []string {
	// 需要过滤掉的标点符号（Unicode码点范围）
	isPunctuation := func(r rune) bool {
		// ASCII标点
		if r >= 0x21 && r <= 0x2F {
			return true // !"#$%&'()*+,-./:
		}
		if r >= 0x3A && r <= 0x40 {
			return true // :;<=>?@:
		}
		if r >= 0x5B && r <= 0x60 {
			return true // [\]^_`:
		}
		if r >= 0x7B && r <= 0x7E {
			return true // {|}~:
		}
		// 全角标点 (FF01-FF0F, FF1A-FF20, FF3B-FF40, FF5B-FF5E)
		if r >= 0xFF01 && r <= 0xFF0F {
			return true
		}
		if r >= 0xFF1A && r <= 0xFF20 {
			return true
		}
		if r >= 0xFF3B && r <= 0xFF40 {
			return true
		}
		if r >= 0xFF5B && r <= 0xFF5E {
			return true
		}
		// 中文标点
		switch r {
		case 0x300A, 0x300B: // 《》
		case 0x3008, 0x3009: // 〈〉
		case 0x300C, 0x300D: // 「」
		case 0x300E, 0x300F: // 『』
		case 0x3010, 0x3011: // 【】
		case 0x2018, 0x2019: // '' (单引号)
		case 0x201C, 0x201D: // "" (双引号)
		case 0x2013, 0x2014: // – (破折号)
		case 0xFF08, 0xFF09: // （） (全角括号)
		case 0xFF3B, 0xFF3D: // （） (全角方括号)
		case 0xFF5B, 0xFF5D: // （） (全角花括号)
		case 0xFF0C: // ， (全角逗号)
		case 0xFF0E: // ． (全角句号)
		case 0xFF1A: // ： (全角冒号)
		case 0xFF1B: // ； (全角分号)
		case 0xFF1F: // ？ (全角问号)
		case 0xFF01: // ！ (全角感叹号)
		case 0xFF5F, 0xFF60: // （） (全角括号)
		case 0x2236: // ∶
			return true
		}
		return false
	}

	isKeepChar := func(r rune) bool {
		// 保留字母、数字、下划线
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
		// 过滤掉标点
		if isPunctuation(r) {
			return false
		}
		return false
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if core.IsDir(path) {
			return nil
		}
		oldName := info.Name()
		ext := filepath.Ext(oldName)
		nameWithoutExt := strings.TrimSuffix(oldName, ext)

		// 逐Unicode字符过滤
		var cleaned []rune
		for _, r := range nameWithoutExt {
			if isKeepChar(r) {
				cleaned = append(cleaned, r)
			}
		}

		result := string(cleaned)

		// 如果清理后文件名为空，给一个默认名
		if result == "" {
			result = "unnamed"
		}

		newName := result + ext
		if newName != oldName {
			newPath := filepath.Dir(path) + "/" + newName
			// 如果新文件名已存在，添加序号
			if _, err := os.Stat(newPath); err == nil {
				baseName := result
				for i := 1; i <= 999; i++ {
					newName = fmt.Sprintf("%s_%d%s", baseName, i, ext)
					newPath = filepath.Dir(path) + "/" + newName
					if _, err := os.Stat(newPath); os.IsNotExist(err) {
						break
					}
				}
			}
			if err := os.Rename(path, newPath); err != nil {
				fmt.Println("清理文件名失败:", oldName, "->", newName, "错误:", err)
			} else {
				fmt.Println("清理文件名成功:", oldName, "->", newName)
			}
		}
		return nil
	})
	return GetFilepathList(dir)
}
