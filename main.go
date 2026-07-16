// OIDC Proxy — 将彩虹聚合登录包装成标准 OIDC 协议的中转站喵
// 配置通过 Web 管理后台修改，数据存储在 SQLite 数据库中喵
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"oidc-proxy/internal/crypto"
	"oidc-proxy/internal/handler"
	"oidc-proxy/internal/middleware"
	"oidc-proxy/internal/rainbow"
	"oidc-proxy/internal/store"

	_ "modernc.org/sqlite"
)

func main() {
	// 解析命令行参数喵
	dbPath := flag.String("db", "data.db", "SQLite 数据库文件路径喵")
	flag.Parse()

	// 初始化 SQLite 数据库喵
	db, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		log.Fatalf("[FATAL] 初始化数据库失败喵: %v\n", err)
	}
	defer db.Close()

	// 初始化 RSA 密钥管理器喵
	keyManager, err := crypto.NewKeyManager("keys")
	if err != nil {
		log.Fatalf("[FATAL] 初始化密钥管理器失败喵: %v\n", err)
	}
	log.Println("[INFO] RSA 密钥管理器就绪喵")

	// 从数据库读配置创建彩虹客户端喵
	rainCfg := db.BuildRainConfig()
	rainClient := rainbow.NewClient(
		rainCfg.AppID, rainCfg.AppKey, rainCfg.ConnectURL,
	)
	log.Println("[INFO] 彩虹 API 客户端就绪喵")

	// 创建 HTTP Handler 喵
	h := handler.NewHandler(db, keyManager, rainClient)

	// ── 路由注册 ──
	mux := http.NewServeMux()

	// OIDC 端点喵
	mux.HandleFunc("/.well-known/openid-configuration", h.HandleDiscovery)
	mux.HandleFunc("/.well-known/jwks.json", h.HandleJWKS)
	mux.HandleFunc("/oidc/authorize", h.HandleAuthorize)
	mux.HandleFunc("/oidc/callback", h.HandleCallback)
	mux.HandleFunc("/oidc/token", h.HandleToken)
	mux.HandleFunc("/oidc/userinfo", h.HandleUserInfo)

	// 静态资源（嵌入的登录图标）喵
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(handler.Assets))))

	// 管理后台认证中间件喵
	adminAuth := middleware.AdminAuth(db)

	// 管理路由喵
	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleLoginPage(w, r)
		case http.MethodPost:
			h.HandleLoginSubmit(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// 需登录的管理路由喵
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin", h.HandleAdminDashboard)
	adminMux.HandleFunc("/admin/logout", h.HandleLogout)
	adminMux.HandleFunc("/admin/settings", h.HandleSaveSettings)
	adminMux.HandleFunc("/admin/clients", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreateClient(w, r)
		case http.MethodDelete:
			h.HandleDeleteClient(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.Handle("/admin", adminAuth(adminMux))
	mux.Handle("/admin/", adminAuth(adminMux))

	// access log 中间件喵
	final := accessLogMW(mux)

	// 启动信息喵
	issuer := db.GetSetting("issuer")
	listen := db.GetSetting("listen")
	log.Printf("[INFO] OIDC Proxy 启动成功喵~")
	log.Printf("[INFO] 数据库: %s", *dbPath)
	log.Printf("[INFO] 监听地址: %s", listen)
	log.Printf("[INFO] Issuer: %s", issuer)
	log.Printf("[INFO] 管理后台: %s/admin", issuer)

	server := &http.Server{
		Addr: listen, Handler: final,
		ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 120 * time.Second,
	}

	log.Printf("[FATAL] 服务器异常终止: %v\n", server.ListenAndServe())
}

// ────────── access log ──────────

func accessLogMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := &rw{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(wr, r)
		log.Printf("[ACCESS] %s %s %d %v", r.Method, r.URL.Path, wr.code, time.Since(start).Round(time.Millisecond))
	})
}

type rw struct {
	http.ResponseWriter
	code int
}

func (w *rw) WriteHeader(c int) { w.code = c; w.ResponseWriter.WriteHeader(c) }
