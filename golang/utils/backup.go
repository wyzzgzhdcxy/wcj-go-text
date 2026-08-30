package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func RestoreGitSshFile(backupDir string) string {
	// 获取用户主目录
	b := core.ReadFileToByte(filepath.Join(backupDir, "backup.meta.json"))
	var logMap map[string]string
	core.JsonToObject(&b, &logMap)
	// 还原每个文件
	for fn, oriFn := range logMap {
		file := filepath.Join(backupDir, fn)
		var err error
		if core.IsDir(file) {
			err = core.CopyDir(file, oriFn)
		} else {
			_, err = core.CopyFile(file, oriFn)
		}
		if err != nil {
			return fmt.Sprintf("还原失败: %s -> %s\n", file, oriFn)
		} else {
			fmt.Printf("成功还原: %s -> %s\n", file, oriFn)
		}
	}
	return ""
}

func GetBackupFileList() []string {
	homeDir := core.GetUserHomeDir()
	var filesToBackup []string
	// 要备份的用户目录下的文件列表
	backFilesNameList := []string{".ssh", ".gitconfig", ".npmrc", ".pnpm-store", ".gradle"}
	for _, fn := range backFilesNameList {
		filesToBackup = append(filesToBackup, filepath.Join(homeDir, fn))
	}
	return filesToBackup
}

func BackupGitSshFile(dir string) string {
	// 备份文件日志，value -源文件路径 key-文件名或目录名
	var logMap = make(map[string]string)
	for _, src := range GetBackupFileList() {
		if _, err := os.Stat(src); os.IsNotExist(err) {
			fmt.Printf("文件不存在，跳过: %s\n", src)
			continue
		}
		var err error
		if core.IsDir(src) {
			err = core.CopyDir(src, filepath.Join(dir, filepath.Base(src)))
		} else {
			_, err = core.CopyFile(src, filepath.Join(dir, filepath.Base(src)))
		}
		logMap[filepath.Base(src)] = src
		if err != nil {
			return fmt.Sprintf("备份 %s 失败: %v\n", src, err)
		} else {
			fmt.Printf("成功备份: %s -> %s\n", src, dir)
		}
	}
	defer func() {
		core.WriteStrToFile(filepath.Join(dir, "backup.meta.json"), core.ToJsonString(logMap))
	}()
	fmt.Printf("Git 密钥备份完成，备份目录: %s\n", dir)
	return ""
}

// BackupWindowsEnv 备份系统和用户环境变量
func BackupWindowsEnv(dir string) string {
	r1 := backupWindowEnvByEnvName(dir, "user")
	if len(r1) == 0 {
		r2 := backupWindowEnvByEnvName(dir, "system")
		return r2
	}
	return r1
}

func backupWindowEnvByEnvName(dir string, envName string) string {
	// 创建带时间戳的备份文件名
	backupFile := filepath.Join(dir, envName+"_env_backup.reg")
	// 执行 reg export 命令
	exportRegStr := `HKEY_CURRENT_USER\Environment`
	if envName == "system" {
		exportRegStr = `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	}
	core.ExecuteCommandReturnByte("reg", "export", exportRegStr, backupFile, "/y")
	fmt.Printf("正在备份%s环境变量到: %s\n", envName, backupFile)
	return ""
}

func RestoreEnvVars(backupDir string) string {
	r1 := RestoreEnvVarsByEnvName(backupDir, "user")
	if len(r1) == 0 {
		r2 := RestoreEnvVarsByEnvName(backupDir, "system")
		return r2
	}
	fmt.Println("注意：某些环境变量可能需要注销或重启后才会生效。")
	return r1
}

func RestoreEnvVarsByEnvName(backupDir string, envName string) string {
	// 默认备份文件路径
	regFile := filepath.Join(backupDir, envName+"_env_backup.reg")
	// 执行reg导入命令
	core.ExecuteCommandReturnByte("reg", "import", regFile)
	fmt.Printf("正在导入注册表文件: %s\n", regFile)
	fmt.Println(envName + "环境变量已成功从注册表文件恢复！")
	return ""
}

func BackUpEnvFiles(backupDir string) string {
	dir := filepath.Join(backupDir, "常用配置文件")
	result := BackupGitSshFile(dir)
	if len(result) == 0 {
		result = BackupWindowsEnv(dir)
		err := core.Tar(filepath.Join(dir), dir+".tar", false)
		if err != nil {
			fmt.Printf("<UNK>: %v\n", err)
		}
		core.DeleteDir(dir)
	}
	return result
}

func RestoreEnvFiles(backupDir string) string {
	dir := filepath.Join(backupDir, "常用配置文件")
	core.UnTar(filepath.Join(backupDir, "常用配置文件.tar"), dir)
	result := RestoreGitSshFile(dir)
	if len(result) == 0 {
		result = RestoreEnvVars(dir)
		core.DeleteDir(dir)
	}
	return result
}
