package theme

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/fx"

	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type urlZipThemeFetcherImpl struct {
	fx.Out
	PropertyScanner PropertyScanner
}

func NewURLZipThemeFetcher(propertyScanner PropertyScanner) ThemeFetcher {
	return &urlZipThemeFetcherImpl{
		PropertyScanner: propertyScanner,
	}
}

func (u *urlZipThemeFetcherImpl) FetchTheme(ctx context.Context, file interface{}) (*dto.ThemeProperty, error) {
	zipURL, ok := file.(string)
	if !ok {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("invalid URL")
	}

	// 验证 URL 格式
	if !strings.HasPrefix(zipURL, "http://") && !strings.HasPrefix(zipURL, "https://") {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("URL must start with http:// or https://")
	}

	// 创建 HTTP 客户端，设置超时
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// 下载文件
	resp, err := client.Get(zipURL)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("failed to download theme: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("failed to download theme: HTTP %d", resp.StatusCode))
	}

	// 从 URL 中提取文件名
	urlParts := strings.Split(zipURL, "/")
	fileName := urlParts[len(urlParts)-1]
	if !strings.HasSuffix(fileName, ".zip") {
		fileName = fileName + ".zip"
	}

	tempDir := os.TempDir()
	diskFilePath := filepath.Join(tempDir, fileName)

	// 如果文件已存在，删除它
	if util.FileIsExisted(diskFilePath) {
		err = os.Remove(diskFilePath)
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("failed to remove existing file")
		}
	}

	// 创建文件
	diskFile, err := os.OpenFile(diskFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("failed to create file")
	}
	defer diskFile.Close()

	// 保存下载的内容
	_, err = io.Copy(diskFile, resp.Body)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("failed to save file")
	}

	// 解压文件
	_, err = util.Unzip(diskFilePath, filepath.Join(tempDir, strings.TrimSuffix(fileName, ".zip")))
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("failed to unzip file")
	}

	dest := filepath.Join(tempDir, strings.TrimSuffix(fileName, ".zip"))

	// 尝试读取主题属性
	themeProperty, err := u.PropertyScanner.ReadThemeProperty(ctx, dest)
	if err == nil && themeProperty != nil {
		return themeProperty, nil
	}

	// 如果根目录没有主题配置，尝试在子目录中查找
	dirEntrys, err := os.ReadDir(dest)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("failed to read directory")
	}

	for _, dirEntry := range dirEntrys {
		if !dirEntry.IsDir() {
			continue
		}
		themeProperty, err = u.PropertyScanner.ReadThemeProperty(ctx, filepath.Join(dest, dirEntry.Name()))
		if err == nil && themeProperty != nil {
			return themeProperty, nil
		}
	}

	return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("theme configuration not found in zip file")
}
