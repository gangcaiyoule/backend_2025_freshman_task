package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // 必须导入 MySQL 驱动
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// 定义请求和响应结构
type RequestBody struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	ConversationID int    `json:"conversation_id"`
}

type ResponseBody struct {
	Result string `json:"result"`
}

// 用户相关结构
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // 加密后的密码，不输出到JSON
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	IsVip     string    `json:"isVip"`
	Role      string    `json:"role"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// 注册请求和响应
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

// 登录请求和响应
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// 获取用户信息响应体
type UserInfoResponse struct {
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	CreateTime time.Time `json:"create_time"`
	Name       string    `json:"name"`
	IsVip      string    `json:"is_vip"`
	Role       string    `json:"role"`
}

// 生成唯一ID
func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 生成会话Token
func generateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 密码加密
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// 验证密码
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func main() {
	//初始化数据库
	initDB()
	//接口
	r := gin.Default()
	r.Use(cors.Default())
	//注册接口
	r.POST("/register", RegisterHandler)
	r.POST("/login", LoginHandler)
	r.POST("/generate", AuthMiddleware(), GenerateHandler())
	r.GET("/getUserHandler", GetUserHandler)
	r.GET("/history", AuthMiddleware(), GetHistoryHandle)
	r.POST("/recharge", AuthMiddleware(), ReCharge)
	r.GET("/getModel", AuthMiddleware(), GetModel)
	r.POST("/logout", AuthMiddleware(), logout)
	r.POST("/newConversation", AuthMiddleware(), NewConversation)
	r.GET("/getConversation", AuthMiddleware(), GetConversation)
	r.GET("/history/:conversation_id", AuthMiddleware(), GetHistory)

	admin := r.Group("/admin", AuthMiddleware(), AdminMiddleware())
	{
		admin.GET("/users", GetAllUsers)
		admin.POST("/users", AddUsers)
		admin.DELETE("/users/:id", DeleteUsers)
	}

	r.Run(":8080")
}

func DeleteUsers(c *gin.Context) {
	userID := c.Param("id")
	//检查ID是否存在
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		log.Printf("查询该ID是否存在失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询该ID是否存在失败"})
		return
	}
	if !exists {
		c.JSON(http.StatusConflict, gin.H{"error": "该ID用户不存在"})
		return
	}
	//查询用户信息
	var user User
	db.QueryRow("SELECT id, email, password, created_time, name, is_vip, role FROM users WHERE id = ?", userID).Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.Name, &user.IsVip, &user.Role)

	_, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("删除该用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除该用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
		"user":    user,
	})

}

func AddUsers(c *gin.Context) {
	var resp User
	err := c.ShouldBindJSON(&resp)
	if err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	//验证邮箱是否已被注册
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", resp.Email).Scan(&exists)
	if err != nil {
		log.Printf("查询该邮箱是否已被注册失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询该邮箱是否已被注册失败"})
		return
	}
	if exists {
		log.Printf("该邮箱已存在: ")
		c.JSON(http.StatusConflict, gin.H{"error": "该邮箱已被注册"})
		return
	}
	//密码加密
	hashPassword, err := hashPassword(resp.Password)
	if err != nil {
		log.Printf("密码哈希加密失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	//生成ID
	userID := generateID()
	_, err = db.Exec("INSERT INTO users (id, email, password, created_time, name, is_vip, role) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, resp.Email, hashPassword, time.Now(), resp.Name, resp.IsVip, resp.Role)
	if err != nil {
		log.Printf("添加用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "添加用户成功"})
}

func GetAllUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, email, created_time, name, is_vip, role FROM users")
	if err != nil {
		log.Printf("查询账号信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询账号信息失败"})
		return
	}
	db.Close()
	var users []UserInfoResponse
	for rows.Next() {
		var resp UserInfoResponse
		if err := rows.Scan(&resp.UserID, &resp.Email, &resp.CreateTime, &resp.Name, &resp.IsVip, &resp.Role); err != nil {
			log.Printf("提取每行账户信息失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提取每行账户信息失败"})
			return
		}
		users = append(users, resp)
	}
	c.JSON(http.StatusOK, gin.H{"users": users})

}

// 获取单个会话记录
func GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	convID := c.Param("conversation_id")
	conversationID, err := strconv.Atoi(convID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id 参数错误"})
		return
	}

	rows, err := db.Query("SELECT role, message, create_time FROM chat_history WHERE user_id = ? AND conversation_id = ? ORDER BY create_time ASC", userID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史失败"})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var role, message string
		var createTime time.Time
		if err := rows.Scan(&role, &message, &createTime); err != nil {
			continue
		}
		history = append(history, gin.H{
			"role":        role,
			"message":     message,
			"create_time": createTime,
		})
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}

// 获取该用户全部会话
func GetConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	rows, err := db.Query("SELECT id, title, create_time FROM conversations WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("获取该用户的会话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取该用户的会话失败"})
		return
	}
	//释放空间
	defer rows.Close()
	//遍历该用户会话
	var convs []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var create_time time.Time
		if err := rows.Scan(&id, &title, &create_time); err != nil {
			continue
		}
		convs = append(convs, gin.H{
			"conversation_id": id,
			"title":           title,
			"create_time":     create_time,
		})

	}
	c.JSON(http.StatusOK, gin.H{"conversations": convs})

}

// 新建会话
func NewConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	now := time.Now()
	title := now.Format("2006-01-02 15:04:05")
	res, err := db.Exec("INSERT INTO conversations (user_id, title, create_time) "+
		"VALUES (?, ?, ?)", userID, title, now)
	if err != nil {
		log.Printf("创建会话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"conversation_id": id, "title": title})
}

// 退出登录
func logout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err != nil {
		log.Printf("logout: token获取失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	_, err = db.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		log.Printf("token从数据库删除失败%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

func GetModel(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	//获取VIP状态
	var is_vip string
	err := db.QueryRow("SELECT is_vip FROM users WHERE id = ?", userID).Scan(&is_vip)
	if err != nil {
		log.Printf("查询VIP状态失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询VIP失败"})
		return
	}
	var model []string
	if is_vip == "true" {
		model = []string{"deepseek-chat", "deepseek-reasoner"}
	} else {
		model = []string{"deepseek-chat"}
	}

	c.JSON(http.StatusOK, gin.H{"availableModel": model})
}

func ReCharge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	//userID := "xDJmhNKE0XtONMFC5ScSyQ=="
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	//成为会员
	_, err := db.Exec("UPDATE users SET is_vip = ? WHERE id = ?", "true", userID)
	if err != nil {
		log.Printf("成为会员失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "成为会员失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "充值成功"})

}

func GetHistoryHandle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	rows, err := db.Query("SELECT role, message, create_time FROM chat_history WHERE user_id = ? ORDER BY create_time ASC ", userID)
	if err != nil {
		log.Printf("获取历史记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}
	defer rows.Close()
	//遍历获取历史记录
	var history []map[string]interface{}
	for rows.Next() {
		var role, message string
		var create_time time.Time
		if err := rows.Scan(&role, &message, &create_time); err != nil {
			continue
		}
		history = append(history, gin.H{
			"role":        role,
			"message":     message,
			"create_time": create_time,
		})
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}

func initDB() {
	var err error
	dsn := "root:zsc060110@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")
}

// 注册接口
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证邮箱是否已被注册
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
	if err != nil {
		log.Printf("错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	if exists {
		log.Printf("用户已存在: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
		return
	}

	// 密码加密
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("密码加密错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 插入新用户
	userID := generateID()
	_, err = db.Exec("INSERT INTO users (id, email, password, created_time, name) VALUES (?, ?, ?, ?, ?)",
		userID, req.Email, hashedPassword, time.Now(), req.Name)
	if err != nil {
		log.Printf("插入用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, RegisterResponse{UserID: userID})
}

// 登录接口
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 查询用户
	var userID, hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", req.Email).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		log.Printf("用户不存在: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	} else if err != nil {
		log.Printf("错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 验证密码
	if !checkPasswordHash(req.Password, hashedPassword) {
		log.Printf("密码错误: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成会话 Token
	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// 保存会话到数据库
	_, err = db.Exec("INSERT INTO sessions (token, user_id, expires_time) VALUES (?, ?, ?)", token, userID, expiresAt)
	if err != nil {
		log.Printf("错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	//设置cookie
	c.SetCookie("session_token", token, 3600*24, "/", "", false, true)
	fmt.Println("Set cookie session_token:", token)
	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"user_id": userID,
	})
}

// 管理员验证中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}
		var role string
		err := db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
		if err != nil || role != "admin" {
			log.Printf("查询用户身份失败: %v", err)
			c.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if token == "" || err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		var userID string
		err = db.QueryRow("SELECT user_id FROM sessions WHERE token = ? AND expires_time > ?", token, time.Now()).Scan(&userID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的 token"})
			c.Abort()
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
			c.Abort()
			return
		}

		// 将 userID 保存到上下文
		c.Set("user_id", userID)
		c.Next()
	}
}

func GenerateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody RequestBody //这时候reqBody还没收到前端传来的数据
		//reqBody接收前端传来的数据
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		// 动态选择模型
		llm, err := openai.New(
			openai.WithModel(reqBody.Model), // req.Model 是前端传来的模型名
			openai.WithToken("sk-81b1a7f9cae9463e850393c4bc73471d"),
			openai.WithBaseURL("https://api.deepseek.com"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "模型初始化失败"})
			return
		}

		log.Printf("用户 %s 请求生成内容", userID.(string))

		// 1. 从数据库取最近 10 条历史记录
		rows, err := db.Query("SELECT role, message FROM chat_history WHERE user_id = ? AND conversation_id = ? ORDER BY create_time ASC LIMIT 10", userID.(string), reqBody.ConversationID)
		if err != nil {
			log.Printf("获取历史记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
			return
		}
		defer rows.Close()

		var history string
		for rows.Next() {
			var role, message string
			if err := rows.Scan(&role, &message); err == nil {
				if role == "user" {
					history += "用户: " + message + "\n"
				} else {
					history += "AI: " + message + "\n"
				}
			}
		}

		// 2. 拼接上下文 + 新的问题
		fullPrompt := history + "用户: " + reqBody.Prompt + "\nAI:"

		ctx := context.Background()
		var result string

		// 流式输出累加
		_, err = llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			fullPrompt,
			llms.WithMaxTokens(500), // 最大生成 500 个 token
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				result += string(chunk)
				return nil
			}),
			llms.WithTemperature(0.8),
			llms.WithStopWords([]string{"\n", "END"}),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("生成失败: %v", err)})
			return
		}

		_, err = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)", userID, reqBody.ConversationID, "user", reqBody.Prompt, time.Now())
		if err != nil {
			log.Printf("保存用户聊天记录失败: %v", err)
		}

		_, err = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)", userID, reqBody.ConversationID, "ai", result, time.Now())
		if err != nil {
			log.Printf("保存AI聊天记录失败: %v", err)
		}

		c.JSON(http.StatusOK, ResponseBody{Result: result})
	}
}

// 获取用户信息接口
func GetUserHandler(c *gin.Context) {
	// 从上下文获取 user_id（由 AuthMiddleware 设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 查询用户信息
	var resp UserInfoResponse
	err := db.QueryRow("SELECT id, email, create_time, is_vip, role FROM users WHERE id = ?", userID.(string)).
		Scan(&resp.UserID, &resp.Email, &resp.CreateTime, &resp.IsVip, &resp.Role)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, resp)
}
