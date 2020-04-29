package main

import (
	"books/models"
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	key := "have_mercy_do_not_attack_my_soft_site"
	//
	//gin.DisableConsoleColor()
	////f, _ := os.Create("gin.log")
	////gin.DefaultWriter = io.MultiWriter(f)
	////ignoresPath := []string{"/login", "/register", "/health", "/password","/book/statistics"}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	//r.Use(func(c *gin.Context) {
	//
	//	path := c.FullPath()
	//	i := 0
	//	f := true
	//	for ; i < len(ignoresPath); i++ {
	//		if ignoresPath[i] == path {
	//			f = false
	//			break
	//		}
	//	}
	//	if f {
	//		auth := c.GetHeader("auth")
	//		if strings.Split(auth, "-")[1] != key {
	//			c.Abort()
	//			return
	//		}
	//	}
	//})


	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{""},
		Timeout:  60 * time.Second,
		Database: "",
		Username: "",
		Password: "",
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatalf("CreateSession failed:%n", err)
	}

	//设置连接池的大小
	session.SetPoolLimit(300)

	if err != nil {
		panic(err)
	}
	defer session.Close()

	chapterDB := session.DB("book").C("xchapter")
	//chapterDB := session.DB("book").C("chapter")
	accountDB := session.DB("book").C("account")
	bookDB := session.DB("book").C("xbiquge")
	rankDB := session.DB("book").C("rank")
	infoDB := session.DB("book").C("info")
	//bookDB := session.DB("book").C("book")
	r.GET("/active/:code", func(c *gin.Context) {
		code := c.Param("code")
		go func() { accountDB.Update(bson.M{"name": code}, bson.M{"$set": bson.M{"state": 1}}) }()
		c.JSON(http.StatusOK, gin.H{"msg": "账号激活成功", "code": 200, "data": ""})
	})
	r.GET("/info", func(c *gin.Context) {
		var info []models.Info
		infoDB.Find(bson.M{}).All(&info)
		c.JSON(http.StatusOK, gin.H{"msg": "", "code": 200, "data": info})
	})
	r.PATCH("/password", func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBind(&user); err != nil {
			panic(err)
		}
		count, err2 := accountDB.Find(bson.M{"name": user.Name, "email": user.EMail}).Count()
		if err2 != nil {
			panic(err2)
		}
		if count > 0 {
			hash := sha1.New()
			io.WriteString(hash, user.PassWord)
			user.PassWord = string(hash.Sum(nil))
			accountDB.Update(bson.M{"name": user.Name}, bson.M{"$set": bson.M{"password": user.PassWord}})
			c.JSON(http.StatusOK, gin.H{"msg": "修改密码成功", "code": 200, "data": ""})
		}

	})
	r.POST("/register", func(c *gin.Context) {
		var regUser models.RegUser
		if err := c.ShouldBind(&regUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "需要认证参数", "code": 400, "data": ""})
			return
		}
		n, e := accountDB.Find(bson.M{"name": regUser.Name}).Count()
		if e != nil {
			return
		}
		if n > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "用户已存在", "code": 400, "data": ""})
			return
		}
		/**
		simple password encry
		*/
		hash := sha1.New()
		io.WriteString(hash, regUser.PassWord)
		regUser.PassWord = string(hash.Sum(nil))
		regUser.State = 0
		e1 := accountDB.Insert(regUser)

		if e1 != nil {
			c.JSON(http.StatusOK, gin.H{"msg": "注册失败,请重试", "code": 200, "data": ""})

			return
		}
		go func() {
			m := gomail.NewMessage()
			m.SetAddressHeader("From", "", "DeerBookAdmin")
			m.SetHeader("To", regUser.EMail)
			m.SetHeader("Subject", "DeerBook Account Active")
			m.SetBody("text/html", " "+regUser.Name+",欢迎使用DeerBook,请点击链接激活账号"+"<a href='https://book.leetomlee.xyz/active/"+regUser.Name+"'>账号激活</a><br/>请加交流联系群,App更新下载在QQ群中.")
			d := gomail.NewDialer("smtp.163.com", 465, "", "")
			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
			if err := d.DialAndSend(m); err != nil {
				log.Fatalln("邮件发送失败" + regUser.EMail)
			}
		}()

		c.JSON(http.StatusOK, gin.H{"msg": "注册成功,请在邮件中激活账号,没收到邮件检查邮箱后加群联系管理", "code": 200, "data": ""})

	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 200, "msg": "ok", "data": ""})

	})
	r.POST("/login", func(c *gin.Context) {
		var user models.LoginUser
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "需要认证参数", "code": 400, "data": ""})
			return
		}
		var users []models.User
		h := sha1.New()
		io.WriteString(h, user.PassWord)
		password := string(h.Sum(nil))
		if e := accountDB.Find(bson.M{"name": user.Name, "password": password}).All(&users); e != nil {
			panic(e)
		}
		if len(users) >= 1 {
			auth := bytes.NewBufferString(users[0].Id.Hex() + "-" + key)
			users[0].Id = ""
			users[0].PassWord = ""
			c.Header("auth", string(auth.Bytes()))
			if users[0].State == 0 {
				c.JSON(http.StatusOK, gin.H{"msg": "登陆失败,账号未激活", "code": 400, "data": ""})
			} else {
				c.JSON(http.StatusOK, gin.H{"msg": "登陆成功", "code": 200, "data": users[0]})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "查无此人", "code": 400, "data": ""})
		}

	})
	books := r.Group("/book")
	{
		books.GET("/detail/:id", func(c *gin.Context) {
			var bookDetail models.BookDetail
			id := c.Param("id")
			if e := bookDB.FindId(id).One(&bookDetail); e != nil {
				panic(e)
			}
			var bks []models.Book
			m := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$ne": bookDetail.Id}, "author": bookDetail.Author}},
				{"$project": bson.M{"book_desc": 0}},
			}
			go func() { bookDB.UpdateId(id, bson.M{"$set": bson.M{"hot": bookDetail.Hot + 1}}) }()
			if e := bookDB.Pipe(m).All(&bks); e != nil {
				panic(e)
			}
			bookDetail.SameAuthorBooks = bks
			count, _ := chapterDB.Find(bson.M{"book_id": bookDetail.Id}).Count()
			bookDetail.Count = count
			c.JSON(http.StatusOK, gin.H{"msg": "", "code": 200, "data": bookDetail})

		})

		books.GET("/shelf", func(c *gin.Context) {
			account := getAccountFromCookie(c, accountDB)
			var books []models.Book
			if account.IdS != nil {
				if e := bookDB.Find(bson.M{"_id": bson.M{"$in": account.IdS}}).All(&books); e != nil {
					panic(e)
				}
			}
			c.JSON(http.StatusOK, gin.H{"msg": "", "code": "200", "data": books})
		})
		books.GET("/hot", func(c *gin.Context) {
			var data []models.Hot
			m := []bson.M{
				{"$sort": bson.M{"hot": -1}},
				{"$limit": 20},
				{"$project": bson.M{"hot": 1, "_id": 1, "book_name": 1}},
			}
			bookDB.Pipe(m).All(&data)
			c.JSON(200, gin.H{
				"data": data,
				"code": 200,
				"msg":  "",
			})
		})
		books.GET("/action/:bookId/:action", func(c *gin.Context) {
			bookId := c.Param("bookId")
			action := c.Param("action")
			account := getAccountFromCookie(c, accountDB)
			if account.IdS != nil {
				i := 0
				f := true
				for ; i < len(account.IdS); i++ {
					if account.IdS[i] == bookId {
						f = false

						break
					}
				}
				if f {
					if action == "add" {
						account.IdS = append(account.IdS, bookId)
						accountDB.UpdateId(account.Id, bson.M{"$set": bson.M{"ids": account.IdS}})
					}
				} else {
					if action == "del" {
						accountDB.UpdateId(account.Id, bson.M{"$set": bson.M{"ids": append(account.IdS[:i], account.IdS[i+1:]...)}})
					}
				}
			} else {
				if action == "add" {
					accountDB.UpsertId(account.Id, bson.M{"$set": bson.M{"ids": []string{bookId}}})
				}
			}

		})
		books.GET("/category", func(c *gin.Context) {
			var data []models.Result

			m := []bson.M{
				{"$group": bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
				{"$project": bson.M{"count": 0, "_id": 1}},
			}
			bookDB.Pipe(m).All(&data)
			var tem []string
			for _, v := range data {
				tem = append(tem, v.Id)
			}
			c.JSON(200, gin.H{
				"data": tem,
				"code": 200,
				"msg":  "",
			})

		})
		books.GET("/category/:category/:page/:size", func(c *gin.Context) {
			page, e1 := strconv.Atoi(c.Param("page"))
			cate := c.Param("category")
			size, e2 := strconv.Atoi(c.Param("size"))

			if e1 != nil {
				panic(e1)
			}
			if e2 != nil {
				panic(e2)
			}
			var datas []models.CateBook
			//db.Table("book").Where("category=?", cate).Select("author,book_name, cover,id").Offset((page - 1) * size).Limit(size).Scan(&datas)
			m := []bson.M{
				{"$match": bson.M{"category": cate}},
				{"$project": bson.M{"author": 1, "book_name": 1, "cover": 1, "_id": 1}},
				{"$skip": (page - 1) * size},
				{"$limit": size},
			}
			bookDB.Pipe(m).All(&datas)
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": datas,
			})
		})
		books.GET("/chapters/:id/:count", func(c *gin.Context) {
			id := c.Param("id")
			count, _ := strconv.Atoi(c.Param("count"))
			var chapters []models.Chapter
			//db.Table("chapter").Where("book_id=?", id).Select("chapter_id,chapter_name").Order("chapter_id asc").Scan(&chapters)
			m := []bson.M{
				{"$match": bson.M{"book_id": id}},
				{"$project": bson.M{"chapter_name": 1, "chapter_id": 1}},
				//{"$sort": bson.M{"chapter_id": 1}},
				{"$skip": count},
			}
			chapterDB.Pipe(m).All(&chapters)
			var temp []models.Chapter

			for _, chapter := range chapters {
				temp = append(temp,
					models.Chapter{ChapterId: chapter.ChapterId, ChapterName: chapter.ChapterName, HasContent: 1})
			}
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": temp,
			})

		})
		books.GET("/chapter/:id", func(c *gin.Context) {
			id := c.Param("id")
			var result models.BookContent
			err = chapterDB.Find(bson.M{"chapter_id": id}).One(&result)
			if err != nil {
				panic(err)
			}
			content := result.Content
			if strings.HasPrefix(content, "http") {
				result.Content = getContent(content)
				go chapterDB.Update(bson.M{"chapter_id": result.Id}, bson.M{"$set": bson.M{"content": result.Content}})
			}
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": result,
			})

		})
		books.GET("/search", func(c *gin.Context) {
			key := c.Query("key")
			page, _ := strconv.Atoi(c.Query("page"))
			size, _ := strconv.Atoi(c.Query("size"))
			var bks []models.Book
			//m := []bson.M{
			//	{"author": bson.M{"$regex": key, "$options": "$im"}},
			//{"$project": bson.M{"book_name": 1, "_id": 1}},
			//}
			var query []bson.M
			//var all []bson.M
			q1 := bson.M{"book_name": bson.M{"$regex": key, "$options": "$i$m"}}
			query = append(query, q1)
			q2 := bson.M{"author": bson.M{"$regex": key, "$options": "$i$m"}}
			//
			query = append(query, q2)

			//q3 := bson.M{"$project": bson.M{"_id": 1, "book_name": 1, "category": 1, "author": 1, "book_desc": 1, "cover": 1}}

			//all = append(all, q3)
			//db.Table("book").Where("book_name LIKE ?", key).Or("author like ?", key).Find(&bks)
			//book.Pipe(m).All(&bks)
			bookDB.Find(bson.M{"$or": query}).Skip((page - 1) * size).Limit(size).All(&bks)
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": bks,
			})
		})
		books.GET("/statistics", func(c *gin.Context) {
			//db.Table("chapter").Count(&chapters)
			n, e := chapterDB.Count()
			if e != nil {
				panic(e)
			}
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": strconv.Itoa(n),
			})

		})
		books.GET("/notice", func(c *gin.Context) {
			var users []models.User
			accountDB.Find(bson.M{}).All(&users)
			var _row map[string]string
			infoDB.Find(bson.M{}).One(&_row)

			for _, user := range users {
				m := gomail.NewMessage()
				m.SetAddressHeader("From", "", "DeerBookAdmin")
				m.SetHeader("To", user.EMail)
				m.SetHeader("Subject", _row["title"])
				m.SetBody("text/html", " "+user.Name+" "+_row["content"])
				d := gomail.NewDialer("smtp.163.com", 465, "", "")
				d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
				if err := d.DialAndSend(m); err != nil {
					log.Fatalln("邮件发送失败" + user.EMail + err.Error())
				}
			}

			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": _row,
			})
		})
		books.GET("/rank/:type", func(c *gin.Context) {
			var _row interface{}

			tpe, _ := strconv.Atoi(c.Param("type"))
			m := []bson.M{
				{"$match": bson.M{"type": tpe}},
				{"$project": bson.M{"_id": 0, "type": 0}},
				//{"$sort": bson.M{"chapter_id": 1}},
			}
			rankDB.Pipe(m).One(&_row)

			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": _row,
			})

		})
		//books.GET("/chapters/sync", func(c *gin.Context) {
		//	command := exec.Command("go", "version")
		//	var buffer bytes.Buffer
		//
		//	command.Stdout = &buffer
		//
		//	run := command.Run()
		//	if run != nil {
		//		panic(run)
		//	}
		//	c.JSON(200, gin.H{
		//		"code": 200,
		//		"msg":  "",
		//		"data": buffer.String(),
		//	})
		//
		//})
	}
	r.Run("0.0.0.0:8081")
}
func getAccountFromCookie(c *gin.Context, accountDB *mgo.Collection) models.Account {
	auth := c.GetHeader("auth")

	id := strings.Split(auth, "-")[0]
	var account models.Account

	if e := accountDB.FindId(bson.ObjectIdHex(id)).One(&account); e != nil {
		panic(e)
	}
	return account

}
func getContent(u string) string {
	// Request the HTML page.
	res, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	content := doc.Find("#content")
	var buf bytes.Buffer

	// Slightly optimized vs calling Each: no single selection object created
	var f func(*html.Node)
	f = func(n *html.Node) {

		if n.Type == html.TextNode {
			// Keep newlines and spaces, like jQuery
			if n.Data != "\n" && n.Data != "," && n.Data != "" {
				buf.WriteString(n.Data)
			}
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil && c.Parent.Parent.Data != "p"; c = c.NextSibling {
				f(c)
			}
		}
	}
	for _, n := range content.Nodes {
		f(n)
	}
	return buf.String()

}
