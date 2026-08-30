package rpc

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"strings"
)

// TwemojiBaseURL Twemoji CDN 基础地址（jsDelivr 镜像 twitter/twemoji）
const TwemojiBaseURL = "https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/256x256"

// EmojiRouteBaseURL emoji-route 服务地址，用于获取带颜色背景的 emoji PNG
const EmojiRouteBaseURL = "https://emoji-route.deno.dev/png"

// TwemojiClient 是对 Twemoji / emoji-route 等 emoji 图片服务的 RPC 客户端封装。
type TwemojiClient struct {
	client *Client
}

// NewTwemojiClient 创建一个新的 Twemoji 客户端。
func NewTwemojiClient() *TwemojiClient {
	return &TwemojiClient{
		client: NewClient(),
	}
}

// BuildTwemojiURL 根据 emoji 字符拼接 Twemoji CDN URL。
func BuildTwemojiURL(emoji string) string {
	codes := make([]string, 0, len(emoji))
	for _, r := range emoji {
		codes = append(codes, fmt.Sprintf("%x", r))
	}
	return fmt.Sprintf("%s/%s.png", TwemojiBaseURL, strings.Join(codes, "-"))
}

// BuildEmojiRouteURL 根据 emoji 字符拼接 emoji-route URL。
func BuildEmojiRouteURL(emoji string) string {
	return fmt.Sprintf("%s/%s", EmojiRouteBaseURL, url.PathEscape(emoji))
}

// DownloadEmojiImage 通用 emoji 图片下载，按 URL 直接拉取字节流。
func (c *TwemojiClient) DownloadEmojiImage(ctx context.Context, imgURL string) ([]byte, error) {
	return c.client.GetBytes(ctx, imgURL, nil)
}

// DownloadEmojiImageAsImage 通用 emoji 图片下载并解码为 image.Image。
func (c *TwemojiClient) DownloadEmojiImageAsImage(ctx context.Context, imgURL string) (image.Image, error) {
	data, err := c.DownloadEmojiImage(ctx, imgURL)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image failed: %w", err)
	}
	return img, nil
}