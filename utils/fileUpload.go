package utils

import (
	"carport-backend/initializers"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandleFileUpload(c *gin.Context, file *multipart.FileHeader) (string, error) {
	// Generate nama file unik
	fileName := uuid.New().String() + filepath.Ext(file.Filename)
	// Lokasi penyimpanan file di server (misalnya, folder 'uploads')
	uploadPath := filepath.Join("uploads", fileName)
	// Simpan file di server
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		return "", err
	}
	// Return URL dari file yang diupload
	return "/uploads/" + fileName, nil
}

func DeleteFile(filePath string) error {
	// Hapus file dari sistem penyimpanan
	if err := os.Remove(filePath[1:]); err != nil {
		return err
	}

	return nil
}

func UploadFileToFirebase(f multipart.FileHeader) (string, error) {
	f.Filename = uuid.New().String() + filepath.Ext(f.Filename)

	bucketName := os.Getenv("FIREBASE_BUCKET")
	ctx := context.Background()
	client, err := initializers.InitFirebaseStorage()
	if err != nil {
		panic(err)
	}

	file, err := f.Open()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	var errs error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wc := client.Bucket(bucketName).Object(f.Filename).NewWriter(ctx)
		if _, err = io.Copy(wc, file); err != nil {
			errs = err
		}
		if err := wc.Close(); err != nil {
			errs = err
		}
		wg.Done()
	}()
	wg.Wait()
	if errs != nil {
		return "", errs
	}

	escapedName := url.QueryEscape(f.Filename)

	// Membuat URL dengan nama file yang telah diubah
	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", bucketName, escapedName)

	return url, nil
}

func DeleteFileFromFirebase(url string) error {
	// Mengambil bagian terakhir yang merupakan nama file
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]

	// Menghilangkan alt=media
	parts = strings.Split(name, "?")
	name = parts[0]

	// Membatalkan escape karakter khusus URL
	unescapedName := strings.ReplaceAll(name, "%20", " ")

	bucketName := os.Getenv("FIREBASE_BUCKET")
	ctx := context.Background()
	client, err := initializers.InitFirebaseStorage()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	var errs error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := client.Bucket(bucketName).Object(unescapedName).Delete(ctx)
		if err != nil {
			errs = err
		}
		wg.Done()
	}()
	wg.Wait()
	if errs != nil {
		return errs
	}

	return nil
}
