package imgproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/external/s3"
)

type signer struct {
	baseURL string
	bucket  string
	key     []byte
	salt    []byte
}

func newImgproxySigner(cfg Config, s3cfg s3.Config) (abstraction.ImgproxySigner, error) {
	return &signer{
		baseURL: cfg.BaseURL,
		bucket:  s3cfg.Bucket,
		key:     cfg.Key,
		salt:    cfg.Salt,
	}, nil
}

func (s *signer) BuildURL(key string, opts abstraction.SignerOptions) string {
	// 1) Build S3 source URL (s3://bucket/key)
	source := fmt.Sprintf("s3://%s/%s", s.bucket, key)

	// 2) Processing options
	var parts []string

	// rs (resize meta): rs:%type:%w:%h
	rt := "fit"
	if opts.Fit != nil && *opts.Fit != "" {
		rt = *opts.Fit
	}
	w := "0"
	if opts.Width != nil && *opts.Width > 0 {
		w = strconv.Itoa(*opts.Width)
	}
	h := "0"
	if opts.Height != nil && *opts.Height > 0 {
		h = strconv.Itoa(*opts.Height)
	}
	parts = append(parts, fmt.Sprintf("rs:%s:%s:%s", rt, w, h))

	if opts.DPR != nil && *opts.DPR > 0 {
		parts = append(parts, "dpr:"+trimFloat(*opts.DPR))
	}
	if opts.Quality != nil && *opts.Quality > 0 {
		parts = append(parts, "q:"+strconv.Itoa(*opts.Quality))
	}
	if opts.Expires != nil && !opts.Expires.IsZero() {
		parts = append(parts, fmt.Sprintf("exp:%d", opts.Expires.Unix()))
	}

	// 3) Source (plain) from S3 URL
	processing := "/" + strings.Join(parts, "/")
	src := "/plain/" + source

	// 4) Extension (@format) — опційно
	suffix := ""
	if opts.Format != nil && *opts.Format != "" {
		ext := strings.TrimPrefix(strings.ToLower(*opts.Format), ".")
		switch ext {
		case "webp", "avif", "jpeg", "jpg", "png":
			if ext == "jpg" {
				ext = "jpeg"
			}
			suffix = "@" + ext
		}
	}

	// 5) Підписуємо повний path (починається з /)
	path := processing + src + suffix
	sig := s.sign(path)

	return s.baseURL + "/" + sig + path
}

func (s *signer) sign(path string) string {
	mac := hmac.New(sha256.New, s.key)
	// важливо: спочатку salt, потім сам path (з провідним "/")
	mac.Write(s.salt)
	mac.Write([]byte(path))
	sum := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}

func trimFloat(f float32) string {
	// красивий вигляд без зайвих нулів (1.5 → "1.5", 2 → "2")
	x := strconv.FormatFloat(float64(f), 'f', -1, 64)
	return strings.TrimRight(strings.TrimRight(x, "0"), ".")
}
