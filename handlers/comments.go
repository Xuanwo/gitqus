package handlers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Xuanwo/gitqus/model"
	"github.com/Xuanwo/gitqus/provider/github"
	"github.com/gin-gonic/gin"
)

const ProviderGithub = "github.com"

type createCommentMeta struct {
	Provider string `uri:"provider" binding:"required"`
	User     string `uri:"user" binding:"required"`
	Repo     string `uri:"repo" binding:"required"`
	Branch   string `uri:"branch" binding:"required"`
}

type createCommentData struct {
	Slug    string `form:"slug" binding:"required"`
	Name    string `form:"name" binding:"required"`
	Email   string `form:"email" binding:"required"`
	Content string `form:"content" binding:"required"`
}

func CreateComment(c *gin.Context) {
	var meta createCommentMeta
	if err := c.ShouldBindUri(&meta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if meta.Provider != ProviderGithub {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only github supported for now"})
		return
	}

	var data createCommentData
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	gc := github.New(meta.User, meta.Repo, meta.Branch, "gitqus", "gitqus@xuanwo.io")
	path := "static" + data.Slug + "comments.json"

	content, err := gc.GetFile(ctx, path)
	if err != nil {
		log.Printf("get file content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get file content failed"})
		return
	}

	comments := make([]*model.Comment, 0)
	if content != nil {
		err = json.Unmarshal(content, &comments)
		if err != nil {
			log.Printf("unmarshall content: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unmarshall content failed"})
			return
		}
	}

	hash := md5.Sum([]byte(data.Email))
	comments = append(comments, &model.Comment{
		Name:      data.Name,
		Email:     hex.EncodeToString(hash[:]),
		Content:   data.Content,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	content, err = json.Marshal(comments)
	if err != nil {
		log.Printf("unmarshall content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unmarshall content failed"})
		return
	}

	message := fmt.Sprintf("%s: Add comment", data.Slug)

	ref, err := gc.CreateOrUpdateFile(ctx, path, message, string(content))
	if err != nil {
		log.Printf("create or update file failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create or update file failed"})
		return
	}

	_, err = gc.CreatePR(ctx, message, ref, meta.Branch)
	if err != nil {
		log.Printf("create pr failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create pr failed"})
		return
	}
}
