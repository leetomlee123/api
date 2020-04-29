package models

import "gopkg.in/mgo.v2/bson"

type Book struct {
	Id             string `bson:"_id"`
	BookName       string `bson:"book_name" json:"Name"`
	Category       string `json:"CName"`
	Rate           int    `bson:"rate"`
	Author         string `bson:"author"`
	UTime          string `bson:"u_time" json:"UTime"`
	BookDesc       string `bson:"book_desc" json:"Desc,omitempty"`
	Status         string `bson:"status" json:"BookStatus"`
	Cover          string `bson:"cover" json:"Img"`
	LastChapter    string `bson:"last_chapter" `
	FirstChapterId string `bson:"first_chapter_id"`
	LastChapterId  string `bson:"last_chapter_id"`
}
type Rank struct {
	Id    string `bson:"id"`
	Cover string `bson:"cover" `
	Name  string `bson:"name"`
}


type RankBook struct {
	Id       string `bson:"_id"`
	BookName string `bson:"book_name" json:"Name"`
	Cover    string `bson:"cover" json:"Img"`
	Author   string `bson:"author"`
	Category string `json:"CName"`
}
type Hot struct {
	Id       string `bson:"_id"`
	BookName string `bson:"book_name" json:"Name"`
	Hot      int64
}
type Info struct {
	Date    string `bson:"date" json:"Date"`
	Title   string `bson:"title" json:"Title"`
	Content string `bson:"content" json:"Content"`
}
type User struct {
	Id       bson.ObjectId `bson:"_id" json:"id,omitempty"`
	Name     string        `form:"name" bson:"name" json:"name" binding:"required"`
	PassWord string        `form:"password" bson:"password" json:"password,omitempty" binding:"required"`
	EMail    string        `form:"email" bson:"email" json:"email"`
	State    int8          `bson:"state"`
}
type LoginUser struct {
	Name     string `form:"name" bson:"name" json:"name" binding:"required"`
	PassWord string `form:"password" bson:"password" json:"password,omitempty" binding:"required"`
}
type RegUser struct {
	Name     string `form:"name" bson:"name" json:"name" binding:"required"`
	PassWord string `form:"password" bson:"password" json:"password" binding:"required"`
	EMail    string `form:"email" bson:"email" json:"email" binding:"required"`
	State    int8   `bson:"state"`
}
type BookDetail struct {
	Id              string  `bson:"_id"`
	BookName        string  `bson:"book_name" json:"Name"`
	Category        string  `json:"CName"`
	Author          string  `bson:"author"`
	Hot             int     `bson:"hot"`
	Rate            float64 `bson:"rate"`
	UTime           string  `bson:"u_time" json:"LastTime"`
	BookDesc        string  `bson:"book_desc" json:"Desc,omitempty"`
	Status          string  `bson:"status" json:"BookStatus"`
	Cover           string  `bson:"cover" json:"Img"`
	LastChapter     string  `bson:"last_chapter" `
	Count           int
	LastChapterId   string `bson:"last_chapter_id"`
	FirstChapterId  string `bson:"first_chapter_id"`
	SameAuthorBooks []Book
}

/**
根据category 分页books
*/
type CateBook struct {
	BookName string `bson:"book_name" json:"bookName"`
	Author   string `json:"author"`
	Cover    string `json:"cover"`
	Id       string `bson:"_id" json:"id"`
}
type Chapter struct {
	ChapterName string `bson:"chapter_name" json:"name"`
	ChapterId   string `bson:"chapter_id" json:"id"`
	HasContent  int    ` json:"hasContent"`
}
type BookContent struct {
	Id      string `bson:"chapter_id" json:"id"`
	Content string `bson:"content" json:"content"`
}
type Result struct {
	Id string `bson:"_id"' `
}

type Account struct {
	Id   bson.ObjectId `bson:"_id"`
	Name string
	IdS  []string `bson:"ids"`
}
