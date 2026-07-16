// Package handler 嵌入静态资源（登录图标）喵
package handler

import "embed"

// Assets 嵌入 picture 目录下的所有图标文件喵
//
//go:embed picture/*.png
var Assets embed.FS
