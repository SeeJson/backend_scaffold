package app

import (
	"mime/multipart"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/SeeJson/backend_scaffold/ginplus"
	"github.com/SeeJson/backend_scaffold/upload"
)

type UploadRequest struct {
	File *multipart.FileHeader `form:"file" json:"file" note:"头像文件"`
}

type UploadResponse struct {
	ginplus.CommonResponse

	URI     string `json:"uri"`
	FullURI string `json:"full_uri"`
}

func NewUploadHandler(u *upload.Uploader) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			logrus.WithField("err", err).Error("read file error")
			ginplus.ResError(c, err)
			return
		}

		//filename := filepath.Base(file.Filename)
		ext := filepath.Ext(file.Filename)
		filename := xid.New().String() + ext
		uri, err := u.SaveGinUploadFile(file, filename)
		if err != nil {
			ginplus.ResError(c, err)
			return
		}

		resp := UploadResponse{
			CommonResponse: ginplus.OkResponse,
			URI:            uri,
			FullURI:        u.FillUrl(uri),
		}
		ginplus.ResSuccess(c, resp)
	}
}
