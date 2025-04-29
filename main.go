package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Article struct {
	Title     string   `json:"title"`
	Slug      string   `json:"slug"`
	Content   string   `json:"content"`
	Platforms []string `json:"platforms"`
	Draft     bool     `json:"draft"`
	Date      string   `json:"date"`
	Lastmod   string   `json:"lastmod"`
	Category  string   `json:"category"`
}

type PlatformConfig struct {
	WeChat struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	} `json:"wechat"`
	Zhihu struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"zhihu"`
	Xiaohongshu struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"xiaohongshu"`
}

var (
	articlesDir = "content/posts"
	configFile  = "config.json"
	db          *sql.DB
)

func main() {
	// 解析命令行参数
	daemon := flag.Bool("d", false, "Run as daemon")
	flag.Parse()

	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/airdrop?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("数据库不可用:", err)
	}

	if *daemon {
		// 后台运行模式
		cmd := exec.Command(os.Args[0])
		cmd.Start()
		fmt.Printf("Server started in background with PID %d\n", cmd.Process.Pid)
		os.Exit(0)
	}

	// 确保文章目录存在
	if err := os.MkdirAll(articlesDir, 0755); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	// 静态文件服务
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// API 路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/articles", handleCreateArticle).Methods("POST")
	api.HandleFunc("/articles/{slug}", handleGetArticle).Methods("GET")
	api.HandleFunc("/articles/{slug}", handleUpdateArticle).Methods("PUT")
	api.HandleFunc("/articles/{slug}", handleDeleteArticle).Methods("DELETE")
	api.HandleFunc("/articles/{slug}/publish", handlePublishArticle).Methods("POST")

	// 管理界面路由
	r.HandleFunc("/admin", handleAdminHome)
	r.HandleFunc("/admin/new", handleNewArticle)
	r.HandleFunc("/admin/edit/{slug}", handleEditArticle)
	r.HandleFunc("/admin/posts", handleArticleList)

	// 根路由处理
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})

	// 启动服务器
	fmt.Println("Server starting on http://localhost:8091")
	log.Fatal(http.ListenAndServe(":8091", r))
}

func handleCreateArticle(w http.ResponseWriter, r *http.Request) {
	var article Article
	if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO articles (title, slug, content, category, draft) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "数据库预处理失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(article.Title, article.Slug, article.Content, article.Category, article.Draft)
	if err != nil {
		http.Error(w, "数据库插入失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(article)
}

func handlePublishArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	// 读取文章
	filePath := filepath.Join(articlesDir, slug+".md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 更新文章状态
	updatedContent := string(content)
	updatedContent = updateFrontMatter(updatedContent, "draft", "false")
	updatedContent = updateFrontMatter(updatedContent, "lastmod", time.Now().Format("2006-01-02"))

	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 同步到其他平台
	var article Article
	if err := json.Unmarshal([]byte(updatedContent), &article); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go syncToPlatforms(article)

	w.WriteHeader(http.StatusOK)
}

func syncToPlatforms(article Article) {
	// 读取平台配置
	config, err := readPlatformConfig()
	if err != nil {
		log.Printf("Error reading platform config: %v", err)
		return
	}

	for _, platform := range article.Platforms {
		switch platform {
		case "wechat":
			syncToWeChat(article, config.WeChat)
		case "zhihu":
			syncToZhihu(article, config.Zhihu)
		case "xiaohongshu":
			syncToXiaohongshu(article, config.Xiaohongshu)
		}
	}
}

func readPlatformConfig() (*PlatformConfig, error) {
	config := &PlatformConfig{}
	file, err := os.ReadFile(configFile)
	if err != nil {
		return config, err
	}
	return config, json.Unmarshal(file, config)
}

// 更新 front matter 中的字段
func updateFrontMatter(content, key, value string) string {
	// 这里需要实现 front matter 的更新逻辑
	return content
}

// 平台同步函数（需要实现具体的平台 API 调用）
func syncToWeChat(article Article, config struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}) {
	// 实现微信公众号同步逻辑
}

func syncToZhihu(article Article, config struct {
	Username string `json:"username"`
	Password string `json:"password"`
}) {
	// 实现知乎同步逻辑
}

func syncToXiaohongshu(article Article, config struct {
	Username string `json:"username"`
	Password string `json:"password"`
}) {
	// 实现小红书同步逻辑
}

// 其他处理函数...
func handleGetArticle(w http.ResponseWriter, r *http.Request)    {}
func handleUpdateArticle(w http.ResponseWriter, r *http.Request) {}
func handleDeleteArticle(w http.ResponseWriter, r *http.Request) {}
func handleAdminHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/admin/index.html")
}
func handleNewArticle(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/admin/new.html")
}
func handleEditArticle(w http.ResponseWriter, r *http.Request) {}
func handleArticleList(w http.ResponseWriter, r *http.Request) {}
