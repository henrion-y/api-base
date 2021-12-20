package gorm

import (
	"context"
	"github.com/henrion-y/api-base/database/gorm"
	_gorm "github.com/jinzhu/gorm"
	"log"
	"testing"
	"time"
)

var dbConfig = gorm.DataBase{
	Driver:   "mysql",
	Host:     "127.0.0.1:3306",
	User:     "root",
	Password: "123456",
	Db:       "test_projects",
	Charset:  "utf8",
}

type User struct {
	ID    int        `json:"id" gorm:"primary_key"`
	Name  string     `json:"name"`
	Age   int        `json:"age"`
	Ctime time.Time  `json:"ctime" gorm:"update_time_stamp"`
	Mtime time.Time  `json:"mtime" gorm:"update_time_stamp"`
	Dtime *time.Time `json:"dtime"`
}

func (t User) TableName() string {
	return "t_user"
}

func getDB() (*_gorm.DB, error) {
	db, err := gorm.NewDbProvider(dbConfig)
	if err != nil {
		log.Fatal("数据库连接失败")
		return nil, err
	}

	return db, nil
}

func TestBaseRepository_Create(t *testing.T) {

	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	db.AutoMigrate(&User{})

	userRepo := NewBaseRepository(db)

	newTime := time.Now()
	user1 := &User{
		Name:  "吕布",
		Age:   28,
		Ctime: newTime,
		Mtime: newTime,
	}

	user2 := &User{
		Name:  "貂蝉",
		Age:   21,
		Ctime: newTime,
		Mtime: newTime,
	}

	{
		err := userRepo.Create(context.Background(), user1)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 3)

		err = userRepo.Create(context.Background(), user2)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("插入记录成功")
	}

}

func TestBaseRepository_FindOne(t *testing.T) {

	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(db)
	user1 := &User{}

	err = userRepo.FindOne(context.Background(), user1, "", "name='吕布'")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(user1)
}

func TestBaseRepository_MultiGet(t *testing.T) {

	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(db)
	user1 := &User{}
	var users []User

	err = userRepo.MultiGet(context.Background(), user1, &users, "", "1=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)

}

func TestBaseRepository_MultiGetByPage(t *testing.T) {
	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(db)
	user1 := &User{}
	var users []User

	err = userRepo.MultiGetByPage(context.Background(), user1, &users, 0, 10, "", "1=1")
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range users {
		t.Log(user)
	}
}

func TestBaseRepository_Update(t *testing.T) {

	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(db)
	user1 := &User{}

	data := map[string]interface{}{
		"age": 30,
	}
	err = userRepo.Update(context.Background(), user1, data, "name='吕布'")
	if err != nil {
		t.Fatal(err)
	}
}

func TestBaseRepository_Count(t *testing.T) {

	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(db)
	user1 := &User{
		Name: "吕布",
		Age:  28,
	}

	count, err := userRepo.Count(context.Background(), user1, "1=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)

}

func TestBaseRepository_Delete(t *testing.T) {
	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(db)
	user1 := &User{
		Name: "吕布",
		Age:  28,
	}
	//err = userRepo.Delete(context.Background(), user1, "name='貂蝉'")
	//if err != nil {
	//	t.Fatal(err)
	//}

	err = userRepo.Delete(context.Background(), user1, "1=1")
	if err != nil {
		t.Fatal(err)
	}

}
