package objectController

import (
	"errors"
	"image"
	"io"
	"mime/multipart"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"jh-oss/internal/apiException"
	"jh-oss/internal/services/objectService"
	"jh-oss/pkg/config"
	"jh-oss/pkg/response"
)

type uploadFileData struct {
	File        *multipart.FileHeader `form:"file" binding:"required"`
	Location    string                `form:"location"`
	DontConvert bool                  `form:"dont_convert"`
	RetainName  bool                  `form:"retain_name"`
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	var data uploadFileData
	if err := c.ShouldBind(&data); err != nil {
		apiException.AbortWithException(c, apiException.ParamError, err)
		return
	}

	fileSize := data.File.Size
	if fileSize > objectService.SizeLimit {
		apiException.AbortWithException(c, apiException.FileSizeExceedError, nil)
		return
	}

	u := uuid.NewV1().String()
	filename := data.File.Filename
	ext := filepath.Ext(filename)             // 获取文件扩展名
	name := filename[:len(filename)-len(ext)] // 获取去掉扩展名的文件名

	// 若不保留文件名，则使用 UUID 作为文件名
	if !data.RetainName {
		name = u
	}

	file, err := data.File.Open()
	if err != nil {
		apiException.AbortWithException(c, apiException.UploadFileError, err)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			zap.L().Warn("文件关闭错误", zap.Error(err))
		}
	}(file)

	// 转换到 WebP
	var reader io.Reader = file
	if !data.DontConvert {
		reader, err = objectService.ConvertToWebP(file)
		ext = ".webp"
		if errors.Is(err, image.ErrFormat) {
			apiException.AbortWithException(c, apiException.FileNotImageError, err)
			return
		}
		if err != nil {
			apiException.AbortWithException(c, apiException.ServerError, err)
			return
		}
	}

	// 上传文件
	objectKey := objectService.GenerateObjectKey(data.Location, name, ext)
	err = objectService.SaveObject(reader, objectKey)
	if err != nil {
		apiException.AbortWithException(c, apiException.ServerError, err)
		return
	}

	zap.L().Info("上传文件成功", zap.String("objectKey", objectKey), zap.String("ip", c.ClientIP()))
	response.JsonSuccessResp(c, gin.H{
		"url": "http://" + config.Config.GetString("oss.domain") + path.Join("/"+config.OSSFolder, objectKey),
	})
}
