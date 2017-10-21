package database

import (
    "testing"
    . "github.com/franela/goblin"
    "github.com/jinzhu/gorm"
    "time"
)

func TestUser(t *testing.T) {
    g := Goblin(t)
    g.Describe("测试 ExistsUser@user.go", func() {
        g.Before(func() {
            Init()
        })
        g.After(func() {
            Init()
        })

        g.It("当存在该用户且密码匹配时返回true", func() {
            g.Assert(ExistsUser("chin", "root")).IsTrue()
        })
        g.It("当存在该用户,但密码不匹配时返回false", func() {
            g.Assert(ExistsUser("chin", "root+1s")).IsFalse()
        })
        g.It("当不存在该用户时应当返回false", func() {
            g.Assert(ExistsUser("hahaha", "+1s")).IsFalse()
        })

    })

    g.Describe("测试user表插入", func() {
        var toBeAddUsers []User
        var db *gorm.DB
        var err error
        var mysqlCount int
        var tmpUser User

        g.Before(func() {
            db, err = ConnectDatabase()
            if err != nil {
                g.Fail("连接mysql错误")
            }
        })

        g.BeforeEach(func() {
            Init()
            toBeAddUsers = []User{
                {UserName:"A", Password:"1", Email:"A@1"},
                {UserName:"B", Password:"2", Email:"B@2"},
                {UserName:"C", Password:"3", Email:"C@3"},
                {UserName:"D", Password:"4", Email:"D@4"},
            }
        })

        g.After(func() {
            defer db.Close()
        })

        g.It(" user_name 不存在的记录正确被插入", func() {
            expectedCount := 1 //init的时候会插入一条root记录
            for _, user := range toBeAddUsers {
                rawPasswd := user.Password
                ok, err := user.DumpToMySQL()
                g.Assert(ok).IsTrue()
                g.Assert(err == nil).IsTrue()
                expectedCount++
                db.Table("users").Count(&mysqlCount)
                g.Assert(expectedCount).Equal(mysqlCount)
                db.Where("user_name = ?", user.UserName).Find(&user).Count(&mysqlCount)
                g.Assert(mysqlCount).Equal(1)
                db.Where("user_name = ?", user.UserName).First(&user)
                g.Assert(user.Password).Eql(toMD5(rawPasswd))
            }
        })

        g.It("插入重复的user_name记录返回失败", func() {
            for _, user := range toBeAddUsers {
                ok, err := user.DumpToMySQL()
                g.Assert(err == nil).IsTrue()
                g.Assert(ok).IsTrue()
                db.Where("user_name = ?", user.UserName).Find(&user).Count(&mysqlCount)
                g.Assert(mysqlCount).Equal(1)
                tmpUser = User{
                    UserName:user.UserName,
                    Password:randomString(3),
                    Email:randomString(10),
                }
                ok, err = tmpUser.DumpToMySQL()
                g.Assert(ok).IsFalse()
                g.Assert(err != nil).IsTrue()
                db.Where("user_name = ?", user.UserName).Find(&user).Count(&mysqlCount)
                g.Assert(mysqlCount).Equal(1)
            }
        })
    })

    g.Describe("测试user表更新", func() {
        var toBeAddUsers []User
        var db *gorm.DB
        var err error
        var mysqlCount int
        var tmpUser User

        g.Before(func() {
            db, err = ConnectDatabase()
            if err != nil {
                g.Fail("连接mysql错误")
            }
        })

        g.BeforeEach(func() {
            Init()
            toBeAddUsers = []User{
                {UserName:"A", Password:"1", Email:"A@1"},
                {UserName:"B", Password:"2", Email:"B@2"},
                {UserName:"C", Password:"3", Email:"C@3"},
                {UserName:"D", Password:"4", Email:"D@4"},
            }
        })

        g.After(func() {
            defer db.Close()
        })

        g.Xit("记录被正确更新", func() {
        //g.It("记录被正确更新", func() {
            for _, user := range toBeAddUsers {
                tmpUser = user
                ok, err := user.DumpToMySQL()
                g.Assert(err == nil).IsTrue()
                g.Assert(ok).IsTrue()
                db.Where("user_name = ?", user.UserName).Find(&user).Count(&mysqlCount)
                g.Assert(mysqlCount).Equal(1)

                oldUpdateTime := user.UpdatedAt
                oldCreateTime := user.CreatedAt
                newPasswd := randomString(30)
                newEmail := randomString(30)
                user.Password = newPasswd
                user.Email = newEmail

                time.Sleep(time.Second)
                user.DumpToMySQL()
                db.Where("user_name =?", tmpUser.UserName).Find(&tmpUser)
                g.Assert(tmpUser.Password).Equal(toMD5(newPasswd))
                g.Assert(tmpUser.Email).Equal(newEmail)
                g.Assert(tmpUser.UpdatedAt.Sub(oldUpdateTime).Seconds() > 0).IsTrue()
                g.Assert(tmpUser.CreatedAt.Sub(oldCreateTime).Seconds() == 0).IsTrue()
            }

        })
    })

    g.Describe("测试 user 数据加载", func() {
        var db *gorm.DB
        var err error
        var users []User

        g.Before(func() {
            db, err = ConnectDatabase()
            if err != nil {
                g.Fail("连接mysql错误")
            }
            Init()
            users = []User{
                {UserName:"A", Password:"1", Email:"A@1"},
                {UserName:"B", Password:"2", Email:"B@2"},
                {UserName:"C", Password:"3", Email:"C@3"},
                {UserName:"D", Password:"4", Email:"C@3"},
            }

            for _, user := range users {
                ok, err := user.DumpToMySQL()
                g.Assert(ok).IsTrue()
                g.Assert(err == nil).IsTrue()
            }
        })

        g.After(func() {
            defer db.Close()
        })

        g.It("记录通过where条件被正确加载", func() {
            for _, user := range users {
                tmpUser, err := new(User).LoadByWhere("user_name = ?", user.UserName)
                g.Assert(err == nil).IsTrue()
                g.Assert(tmpUser.UserName).Equal(user.UserName)
                g.Assert(tmpUser.Password).Equal(toMD5(user.Password))
                g.Assert(tmpUser.Email).Equal(user.Email)
            }
        })

        g.It("记录通主键被正确加载", func() {
            for id, user := range users {
                tmpUser, err := new(User).LoadByKey(id + 2) // 因为已经存在root用户，所以第一个用户的id是2
                g.Assert(err == nil).IsTrue()
                g.Assert(tmpUser.UserName).Equal(user.UserName)
                g.Assert(tmpUser.Password).Equal(toMD5(user.Password))
                g.Assert(tmpUser.Email).Equal(user.Email)
            }
        })

        g.It("记录通过多个where条件被正确加载", func() {
            for id, user := range users {
                tmpUser, err := new(User).LoadByWhere(
                    "id = ? and user_name = ? and email = ?",
                    id + 2, user.UserName, user.Email,
                )
                g.Assert(err == nil).IsTrue()
                g.Assert(tmpUser.UserName).Equal(user.UserName)
                g.Assert(tmpUser.Password).Equal(toMD5(user.Password))
                g.Assert(tmpUser.Email).Equal(user.Email)
            }
        })

        g.It("当存在多于一条记录满足where条件时无法实例化，返回异常且对象为nil", func() {
            tmpUser, err := new(User).LoadByWhere("email = ?", users[2].Email)
            g.Assert(tmpUser == nil).IsTrue()
            g.Assert(err.Error()).Equal("存在多条满足条件的记录，无法实例化")
        })

        g.It("当存在零条记录满足where条件时无法实例化，返回异常且对象为nil", func() {
            tmpUser, err := new(User).LoadByWhere("email = ?", "fuck@shit")
            g.Assert(tmpUser == nil).IsTrue()
            g.Assert(err.Error()).Equal("不存在满足条件的记录，无法实例化")
        })

    })
}