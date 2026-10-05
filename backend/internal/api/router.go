package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.gdys01.cn/gdys01/English-media-learner/backend/ent"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/dictionary"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/media"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/playback"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/subtitle"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	media      *media.Service
	playback   *playback.Service
	subtitle   *subtitle.Service
	dictionary *dictionary.Service
}

func NewRouter(mediaService *media.Service, playbackService *playback.Service, subtitleService *subtitle.Service, dictionaryService *dictionary.Service, frontendDist string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	_ = router.SetTrustedProxies(nil)
	handler := Handler{media: mediaService, playback: playbackService, subtitle: subtitleService, dictionary: dictionaryService}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})
	api := router.Group("/api/v1")
	api.POST("/media/scan", handler.scan)
	api.GET("/media", handler.list)
	api.GET("/media/:id", handler.get)
	api.POST("/media/:id/probe", handler.reprobe)
	api.GET("/media/:id/playback", handler.getPlayback)
	api.GET("/media/:id/playback/:filename", handler.playbackFile)
	api.GET("/media/:id/stream", handler.stream)
	api.GET("/media/:id/subtitles", handler.listSubtitles)
	api.POST("/subtitle-tracks/:trackId/process", handler.processSubtitle)
	api.GET("/subtitle-tracks/:trackId/cues", handler.subtitleCues)
	api.GET("/media/:id/learning-cues", handler.learningCues)
	api.GET("/dictionary/:word", handler.pronunciation)
	if frontendDist != "" {
		serveFrontend(router, frontendDist)
	}
	return router
}

func (h Handler) learningCues(c *gin.Context) {
	mediaID, ok := parseID(c)
	if !ok {
		return
	}
	englishTrackID, ok := parsePositiveQuery(c, "englishTrackId")
	if !ok {
		return
	}
	chineseTrackID, ok := parsePositiveQuery(c, "chineseTrackId")
	if !ok {
		return
	}
	cues, err := h.subtitle.LearningCues(c.Request.Context(), mediaID, englishTrackID, chineseTrackID)
	if err != nil {
		respondEntError(c, err)
		return
	}
	c.JSON(http.StatusOK, cues)
}

func (h Handler) pronunciation(c *gin.Context) {
	result, err := h.dictionary.Lookup(c.Request.Context(), c.Param("word"))
	if err != nil {
		if strings.Contains(err.Error(), "invalid English word") {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h Handler) listSubtitles(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	tracks, err := h.subtitle.ListTracks(c.Request.Context(), id)
	if err != nil {
		respondEntError(c, err)
		return
	}
	c.JSON(http.StatusOK, tracks)
}

func (h Handler) processSubtitle(c *gin.Context) {
	trackID, ok := parsePositiveParam(c, "trackId")
	if !ok {
		return
	}
	result, err := h.subtitle.Process(c.Request.Context(), trackID)
	if err != nil {
		respondEntError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h Handler) subtitleCues(c *gin.Context) {
	trackID, ok := parsePositiveParam(c, "trackId")
	if !ok {
		return
	}
	cues, err := h.subtitle.Cues(c.Request.Context(), trackID)
	if err != nil {
		respondEntError(c, err)
		return
	}
	c.JSON(http.StatusOK, cues)
}

func serveFrontend(router *gin.Engine, dist string) {
	indexPath := filepath.Join(dist, "index.html")
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			respondError(c, http.StatusNotFound, fmt.Errorf("API route not found"))
			return
		}
		relative := strings.TrimPrefix(filepath.Clean(c.Request.URL.Path), string(filepath.Separator))
		candidate := filepath.Join(dist, relative)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			c.File(candidate)
			return
		}
		if _, err := os.Stat(indexPath); err != nil {
			respondError(c, http.StatusNotFound, fmt.Errorf("frontend build not found"))
			return
		}
		c.File(indexPath)
	})
}

func (h Handler) scan(c *gin.Context) {
	result, err := h.media.Scan(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": result.Added, "updated": result.Updated, "unchanged": result.Unchanged, "failed": result.Failed})
}

func (h Handler) list(c *gin.Context) {
	items, err := h.media.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h Handler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.media.Get(c.Request.Context(), id)
	if err != nil {
		respondEntError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h Handler) reprobe(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.media.Reprobe(c.Request.Context(), id)
	if err != nil {
		respondEntError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h Handler) getPlayback(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var requestedAudio *int
	if value := c.Query("audioStream"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			respondError(c, http.StatusBadRequest, fmt.Errorf("invalid audioStream"))
			return
		}
		requestedAudio = &parsed
	}
	info, err := h.playback.Get(c.Request.Context(), id, requestedAudio)
	if err != nil {
		respondEntError(c, err)
		return
	}
	if info.Status == "unsupported" {
		c.JSON(http.StatusUnprocessableEntity, info)
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h Handler) stream(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	path, entity, err := h.media.MediaPath(c.Request.Context(), id)
	if err != nil {
		respondEntError(c, err)
		return
	}
	serveMediaFile(c, path, filepath.Base(entity.Path))
}

func (h Handler) playbackFile(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	path, err := h.playback.DerivedPath(id, c.Param("filename"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			respondError(c, http.StatusNotFound, fmt.Errorf("playback file not found"))
			return
		}
		respondError(c, http.StatusBadRequest, err)
		return
	}
	serveMediaFile(c, path, filepath.Base(path))
}

func serveMediaFile(c *gin.Context, path, name string) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			respondError(c, http.StatusNotFound, fmt.Errorf("media file not found"))
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.Header("Accept-Ranges", "bytes")
	http.ServeContent(c.Writer, c.Request, name, info.ModTime(), file)
}

func parseID(c *gin.Context) (int, bool) {
	return parsePositiveParam(c, "id")
}

func parsePositiveParam(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, fmt.Errorf("invalid %s", name))
		return 0, false
	}
	return id, true
}

func parsePositiveQuery(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Query(name))
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, fmt.Errorf("invalid %s", name))
		return 0, false
	}
	return id, true
}

func respondEntError(c *gin.Context, err error) {
	if ent.IsNotFound(err) {
		respondError(c, http.StatusNotFound, fmt.Errorf("media not found"))
		return
	}
	respondError(c, http.StatusInternalServerError, err)
}

func respondError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}
