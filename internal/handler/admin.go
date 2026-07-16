// Package handler 提供 Web 管理后台（客户端管理 + 系统设置）喵
package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const adminSessionDur = 2 * time.Hour

// ────────── 登录页面 ──────────

const loginHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>OIDC Proxy - 登录</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f0f2f5;display:flex;justify-content:center;align-items:center;min-height:100vh}
.box{background:#fff;padding:40px;border-radius:12px;box-shadow:0 2px 12px rgba(0,0,0,.08);width:100%;max-width:400px}
h1{text-align:center;color:#1a1a2e;margin-bottom:8px;font-size:24px}
.sub{text-align:center;color:#888;margin-bottom:32px;font-size:14px}
.fg{margin-bottom:20px}
.fg label{display:block;margin-bottom:6px;color:#333;font-weight:500;font-size:14px}
.fg input{width:100%;padding:10px 14px;border:1px solid #d9d9d9;border-radius:6px;font-size:14px}
.fg input:focus{outline:none;border-color:#4a6cf7}
.btn{width:100%;padding:10px;background:#4a6cf7;color:#fff;border:none;border-radius:6px;font-size:15px;cursor:pointer}
.btn:hover{background:#3b5de7}
.err{color:#e74c3c;text-align:center;margin-bottom:16px;font-size:14px}
</style></head>
<body><div class="box">
<h1>OIDC Proxy 管理后台</h1>
<p class="sub">请使用管理员账号登录</p>
{{ERROR}}
<form method="POST" action="/admin/login">
<div class="fg"><label>用户名</label><input type="text" name="username" required></div>
<div class="fg"><label>密码</label><input type="password" name="password" required></div>
<button type="submit" class="btn">登 录</button>
</form></div></body></html>`

// ────────── 后台公共顶部+样式 ──────────

const adminHead = `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>OIDC Proxy - 管理后台</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f0f2f5;min-height:100vh}
.header{background:#1a1a2e;color:#fff;padding:0 24px;height:56px;display:flex;align-items:center;justify-content:space-between}
.header h2{font-size:18px;font-weight:500}
.header nav{display:flex;gap:20px;align-items:center}
.header nav a{color:#aab;text-decoration:none;font-size:14px;transition:color .2s}
.header nav a:hover,.header nav a.active{color:#fff}
.container{max-width:1100px;margin:0 auto;padding:24px}
.card{background:#fff;border-radius:10px;box-shadow:0 1px 6px rgba(0,0,0,.06);margin-bottom:24px;overflow:hidden}
.card-h{padding:16px 20px;border-bottom:1px solid #f0f0f0;font-size:16px;font-weight:600;color:#1a1a2e;display:flex;align-items:center;justify-content:space-between}
.card-b{padding:20px}
table{width:100%;border-collapse:collapse}
th,td{text-align:left;padding:10px 14px;border-bottom:1px solid #f0f0f0;font-size:14px}
th{background:#fafafa;font-weight:600;color:#555}
tr:hover{background:#fafafa}
.btn-sm{padding:4px 12px;border-radius:4px;font-size:13px;cursor:pointer;border:none}
.btn-sm:hover{opacity:.85}
.btn-d{background:#e74c3c;color:#fff}
.btn-p{background:#4a6cf7;color:#fff}
.btn-s{background:#27ae60;color:#fff}
.badge{display:inline-block;padding:2px 8px;border-radius:4px;font-size:12px}
.badge-g{background:#e6f7e6;color:#389e0d}
.badge-r{background:#fff1f0;color:#cf1322}
.status-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:16px}
.si{background:#fafafa;padding:16px;border-radius:8px}
.si .lbl{font-size:12px;color:#888;margin-bottom:4px}
.si .val{font-size:18px;font-weight:600;color:#1a1a2e}
.clip{color:#4a6cf7;cursor:pointer;font-size:12px;margin-left:8px}
.clip:hover{text-decoration:underline}
.modal{display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,.45);z-index:100;justify-content:center;align-items:center}
.modal.show{display:flex}
.modal-c{background:#fff;border-radius:10px;padding:24px;width:90%;max-width:500px}
.modal-c h3{margin-bottom:16px}
.fg{margin-bottom:14px}
.fg label{display:block;margin-bottom:4px;color:#333;font-size:13px}
.fg input,.fg select{width:100%;padding:8px 12px;border:1px solid #d9d9d9;border-radius:4px;font-size:14px}
.fg input:focus{outline:none;border-color:#4a6cf7}
.btn-row{display:flex;gap:8px;justify-content:flex-end;margin-top:8px}
.copy-box{margin:8px 0;padding:8px;background:#f5f5f5;border-radius:4px;word-break:break-all;font-family:monospace;font-size:13px}
.tabs{display:flex;gap:0;margin-bottom:24px}
.tabs a{padding:10px 20px;background:#e8e8e8;color:#555;text-decoration:none;font-size:14px;border-radius:6px 6px 0 0}
.tabs a.active{background:#fff;color:#1a1a2e;font-weight:600}
.msg{text-align:center;padding:12px;border-radius:6px;margin-bottom:16px;font-size:14px}
.msg-ok{background:#e6f7e6;color:#389e0d}
.msg-err{background:#fff1f0;color:#cf1322}
</style></head><body>
<div class="header"><h2>OIDC Proxy 管理后台</h2><nav><a href="/admin?tab=dashboard" class="{{DASH_ACTIVE}}">仪表盘</a><a href="/admin?tab=clients" class="{{CLI_ACTIVE}}">客户端</a><a href="/admin?tab=settings" class="{{SET_ACTIVE}}">系统设置</a><a href="/admin/logout">退出</a></nav></div>
<div class="container">`

const adminFoot = `</div></body></html>`

// ────────── 路由处理 ──────────

// HandleLoginPage GET /admin/login 喵
func (h *Handler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	html := strings.Replace(loginHTML, "{{ERROR}}", "", 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// HandleLoginSubmit POST /admin/login 喵
func (h *Handler) HandleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	cfgUser := h.store.GetSetting("admin_username")
	cfgPass := h.store.GetSetting("admin_password")

	if username != cfgUser || password != cfgPass {
		h.logf("WARN", "登录失败: user=%s", username)
		html := strings.Replace(loginHTML, "{{ERROR}}", `<p class="err">用户名或密码错误</p>`, 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
		return
	}

	token := generateRandom(16)
	exp := time.Now().Add(adminSessionDur)
	h.store.SetSession(token, exp)
	http.SetCookie(w, &http.Cookie{
		Name: "admin_session", Value: token, Path: "/admin",
		Expires: exp, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})

	h.logf("INFO", "管理员登录成功: %s", username)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// HandleAdminDashboard GET /admin 喵（含 tab 切换：仪表盘/客户端/设置）
func (h *Handler) HandleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tab := r.URL.Query().Get("tab")

	head := adminHead
	head = strings.Replace(head, "{{DASH_ACTIVE}}", activeClass(tab, "dashboard"), 1)
	head = strings.Replace(head, "{{CLI_ACTIVE}}", activeClass(tab, "clients"), 1)
	head = strings.Replace(head, "{{SET_ACTIVE}}", activeClass(tab, "settings"), 1)

	var body string
	switch tab {
	case "clients":
		body = h.renderClients()
	case "settings":
		body = h.renderSettings()
	default:
		body = h.renderDashboard()
	}

	fmt.Fprint(w, head+body+adminFoot)
}

func activeClass(tab, name string) string {
	if (tab == "" && name == "dashboard") || tab == name {
		return "active"
	}
	return ""
}

// ────────── 仪表盘 ──────────

func (h *Handler) renderDashboard() string {
	uptime := time.Since(h.startTime)
	uptimeStr := fmt.Sprintf("%d天%d时%d分", int(uptime.Hours()/24), int(uptime.Hours())%24, int(uptime.Minutes())%60)

	rainOK := h.rainClient.CheckConnectivity()
	rainBadge := `<span class="badge badge-g">正常</span>`
	if !rainOK {
		rainBadge = `<span class="badge badge-r">异常</span>`
	}

	clients, _ := h.store.ListClients()

	s := fmt.Sprintf(`<div class="card"><div class="card-h">运行状态</div><div class="card-b">
<div class="status-grid">
<div class="si"><div class="lbl">服务启动时间</div><div class="val" style="font-size:14px">%s</div></div>
<div class="si"><div class="lbl">运行时长</div><div class="val" style="font-size:14px">%s</div></div>
<div class="si"><div class="lbl">已注册客户端</div><div class="val">%d</div></div>
<div class="si"><div class="lbl">彩虹 API 连通性</div><div class="val" style="font-size:14px">%s</div></div>
</div></div></div>`,
		h.startTime.Format("2006-01-02 15:04:05"),
		uptimeStr,
		len(clients),
		rainBadge,
	)

	// 登录日志喵
	logs, _ := h.store.ListLogs()
	s += `<div class="card"><div class="card-h">最近登录日志</div><div class="card-b">`
	if len(logs) == 0 {
		s += `<p style="color:#999;text-align:center;padding:20px">暂无登录记录喵~</p>`
	} else {
		s += `<table><tr><th>时间</th><th>应用</th><th>平台</th><th>用户</th></tr>`
		for _, l := range logs {
			s += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, l.Time, l.Client, l.Platform, l.Nickname)
		}
		s += `</table>`
	}
	s += `</div></div>`

	// 弹窗喵
	s += modalsHTML()
	return s
}

// ────────── 客户端管理 ──────────

func (h *Handler) renderClients() string {
	clients, _ := h.store.ListClients()

	s := `<div class="card"><div class="card-h"><span>客户端管理</span><button class="btn-sm btn-p" onclick="showCreate()">+ 创建客户端</button></div><div class="card-b">`
	if len(clients) == 0 {
		s += `<p style="color:#999;text-align:center;padding:20px">暂无客户端，点击右上角创建喵~</p>`
	} else {
		s += `<table><tr><th>名称</th><th>Client ID</th><th>回调地址</th><th>创建时间</th><th>操作</th></tr>`
		for _, c := range clients {
			sid := c.ClientID
			if len(sid) > 16 {
				sid = sid[:16] + "..."
			}
			s += fmt.Sprintf(`<tr><td>%s</td><td><code>%s</code><span class="clip" onclick="viewSecret('%s','%s')">查看密钥</span></td><td style="font-size:13px;max-width:260px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">%s</td><td>%s</td><td><button class="btn-sm btn-d" onclick="delClient('%s')">删除</button></td></tr>`,
				c.Name, sid, c.ClientID, c.ClientSecret, c.RedirectURI, c.CreatedAt, c.ClientID)
		}
		s += `</table>`
	}
	s += `</div></div>` + modalsHTML()
	return s
}

// ────────── 系统设置 ──────────

func (h *Handler) renderSettings() string {
	sets := h.store.GetAllSettings()
	v := func(k string) string { return sets[k] }

	s := `<div class="card"><div class="card-h">系统设置</div><div class="card-b">`
	s += `<p style="color:#888;font-size:13px;margin-bottom:16px">修改后需要重启服务才能生效喵</p>`

	s += fmt.Sprintf(`<form method="POST" action="/admin/settings">
<div class="fg"><label>Issuer (OIDC 签发者 URL)</label><input type="text" name="issuer" value="%s" placeholder="https://oidc.example.com"></div>
<div class="fg"><label>监听地址</label><input type="text" name="listen" value="%s" placeholder=":8443"></div>
<hr style="margin:20px 0;border-color:#f0f0f0">
<h4 style="margin-bottom:14px">管理员账号</h4>
<div class="fg"><label>用户名</label><input type="text" name="admin_username" value="%s"></div>
<div class="fg"><label>密码</label><input type="password" name="admin_password" value="%s"></div>
<hr style="margin:20px 0;border-color:#f0f0f0">
<h4 style="margin-bottom:14px">彩虹聚合登录</h4>
<div class="fg"><label>APP ID</label><input type="text" name="rainbow_app_id" value="%s"></div>
<div class="fg"><label>APP KEY</label><input type="password" name="rainbow_app_key" value="%s"></div>
<div class="fg"><label>Connect URL</label><input type="text" name="rainbow_connect_url" value="%s" placeholder="https://login.mapay.cn/connect.php"></div>
<div class="btn-row"><button type="submit" class="btn-sm btn-s">保存设置</button></div>
</form>`,
		v("issuer"), v("listen"),
		v("admin_username"), v("admin_password"),
		v("rainbow_app_id"), v("rainbow_app_key"),
		v("rainbow_connect_url"),
	)

	s += `</div></div>`
	return s
}

// HandleSaveSettings POST /admin/settings 保存系统设置喵
func (h *Handler) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?tab=settings&saved=0", http.StatusFound)
		return
	}

	keys := []string{
		"issuer", "listen",
		"admin_username", "admin_password",
		"rainbow_app_id", "rainbow_app_key",
		"rainbow_connect_url",
	}
	for _, k := range keys {
		v := strings.TrimSpace(r.FormValue(k))
		_ = h.store.SetSetting(k, v)
	}

	// 重新创建彩虹客户端（因为配置可能变了）喵
	rc := h.store.BuildRainConfig()
	h.rainClient.UpdateConfig(rc)

	h.logf("INFO", "系统设置已更新")
	http.Redirect(w, r, "/admin?tab=settings&saved=1", http.StatusFound)
}

// ────────── 客户端 CRUD API ──────────

// HandleCreateClient POST /admin/clients 喵
func (h *Handler) HandleCreateClient(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	uri := strings.TrimSpace(r.FormValue("redirect_uri"))
	if name == "" || uri == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "名称和回调地址不能为空"})
		return
	}
	client, err := h.store.CreateClient(name, uri)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.logf("INFO", "创建客户端: %s (%s)", client.Name, client.ClientID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "client": client})
}

// HandleDeleteClient DELETE /admin/clients 喵
func (h *Handler) HandleDeleteClient(w http.ResponseWriter, r *http.Request) {
	cid := r.URL.Query().Get("client_id")
	if cid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 client_id"})
		return
	}
	if err := h.store.DeleteClient(cid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.logf("INFO", "删除客户端: %s", cid)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// HandleLogout GET /admin/logout 喵
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("admin_session"); err == nil && c.Value != "" {
		h.store.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "admin_session", Value: "", Path: "/admin", MaxAge: -1})
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// ────────── 弹窗 + JS 喵 ──────────

func modalsHTML() string {
	return `<div class="modal" id="createModal"><div class="modal-c">
<h3>创建 OIDC 客户端</h3>
<div class="fg"><label>应用名称</label><input type="text" id="newName" placeholder="例如：NextCloud"></div>
<div class="fg"><label>回调地址</label><input type="text" id="newRedirect" placeholder="https://example.com/oauth/callback"></div>
<div class="btn-row">
<button class="btn-sm" style="background:#ddd;color:#333" onclick="hideCreate()">取消</button>
<button class="btn-sm btn-s" onclick="doCreate()">确认创建</button>
</div></div></div>

<div class="modal" id="secretModal"><div class="modal-c">
<h3>客户端密钥</h3>
<div class="fg"><label>Client ID</label><div class="copy-box" id="sID"></div></div>
<div class="fg"><label>Client Secret</label><div class="copy-box" id="sSec"></div></div>
<div class="btn-row"><button class="btn-sm" style="background:#ddd;color:#333" onclick="hideSecret()">关闭</button></div>
</div></div>

<script>
function showCreate(){document.getElementById('createModal').classList.add('show')}
function hideCreate(){document.getElementById('createModal').classList.remove('show')}
function hideSecret(){document.getElementById('secretModal').classList.remove('show')}
function viewSecret(cid,sec){document.getElementById('sID').textContent=cid;document.getElementById('sSec').textContent=sec;document.getElementById('secretModal').classList.add('show')}
function doCreate(){
var n=document.getElementById('newName').value.trim();
var r=document.getElementById('newRedirect').value.trim();
if(!n||!r){alert('请填写完整信息');return}
fetch('/admin/clients',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:'name='+encodeURIComponent(n)+'&redirect_uri='+encodeURIComponent(r)})
.then(function(r){return r.json()}).then(function(d){if(d.ok)location.reload();else alert(d.error)}).catch(function(e){alert('失败: '+e)})
}
function delClient(cid){
if(!confirm('确认删除此客户端？'))return;
fetch('/admin/clients?client_id='+encodeURIComponent(cid),{method:'DELETE'})
.then(function(r){return r.json()}).then(function(d){if(d.ok)location.reload();else alert(d.error)}).catch(function(e){alert('失败: '+e)})
}
</script>`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
