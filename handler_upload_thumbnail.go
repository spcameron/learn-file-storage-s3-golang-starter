package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	video, err := cfg.db.GetVideo(videoID)
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User does not have access to this video", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not determine file type from request", err)
		return
	}
	if (mediaType != "image/jpeg") && (mediaType != "image/png") {
		respondWithError(w, http.StatusBadRequest, "Only .png or .jpeg images are allowed", err)
		return
	}

	ext, err := mime.ExtensionsByType(mediaType)

	imgFileExtension := ext[0]

	b := make([]byte, 32)
	rand.Read(b)
	randString := base64.RawURLEncoding.EncodeToString(b)

	imgFilename := fmt.Sprintf("%s%s", randString, imgFileExtension)
	imgFilepath := filepath.Join(cfg.assetsRoot, imgFilename)
	imgFile, err := os.Create(imgFilepath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create file", err)
		return
	}
	defer imgFile.Close()
	io.Copy(imgFile, file)

	imgURL := fmt.Sprintf("http://localhost:%s/assets/%s%s",
		cfg.port,
		randString,
		imgFileExtension)

	video.ThumbnailURL = &imgURL

	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
