package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "database/sql"
	_ "github.com/go-sql-driver/mysql"
	"crypto/md5"
    "encoding/hex"
	"os"
	"io"
	"io/ioutil"
	"time"
	 "strconv"
)


const (
	UPLOAD_DIR = "./uploads"
	DOMAIN = "http://127.0.0.1:8081"
)


type Res struct {
	Code    int         `json:"code"`
	Data    []string `json:"data"`
	Message string      `json:"message"`
}

type Res1 struct {
	Code    int         `json:"code"`
	Data   map[string]string `json:"data"`
	Message string      `json:"msg"`
}

type Info struct{
	id int
	info string
	title string
	img string
	detail string
	addtime string
}

type Res2 struct{
	data []Info
} 

func main() {
    //第一个参数是接口名，第二个参数 http handle func
    http.HandleFunc("/login", login)
     http.HandleFunc("/imgupload", imgupload)
     http.HandleFunc("/img", imgdir)
     http.HandleFunc("/cover", cover)
      http.HandleFunc("/list", list)
    //服务器要监听的主机地址和端口号
    http.ListenAndServe("127.0.0.1:8081", nil)
}

func login(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*") 
	var pas string
	req.ParseForm()
	if req.Form["username"] ==nil || req.Form["password"] ==nil{
		res, _ := json.Marshal(Res{0,[]string{},"登录失败"})
		fmt.Fprint(rw, string(res))
	    return
	}
	username :=req.Form["username"][0]
	password :=req.Form["password"][0]

	if username != "admin"{
		res, _ := json.Marshal(Res{0,[]string{},"登录失败"})
		fmt.Fprint(rw, string(res))
	    return
	} 
	password = md5V(password)
	DB, _ := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test")
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	DB.QueryRow("select password from admin where username ='admin'").Scan(&pas)
	// rows, e := DB.Query("select * from admin where username ='admin'")
	// fmt.Fprint(rw, pas)
	if password != pas{
		res, _ := json.Marshal(Res{0,[]string{},"密码错误"})
		fmt.Fprint(rw, string(res))
	    return
	}else{
		res, _ := json.Marshal(Res{200,[]string{},"登录成功"})
		fmt.Fprint(rw, string(res))
	    return
	}

}


func imgupload(rw http.ResponseWriter, req *http.Request){
	rw.Header().Set("Access-Control-Allow-Origin", "*") 
	req.ParseForm()
	var res Res1
	now := time.Now().Unix()
	tm := time.Unix(now, 0)
	time_now :=tm.Format("2006-01-02 03:04:05")
	time_now = md5V(time_now)
	f, _, err := req.FormFile("file")
		if err != nil {
			http.Error(rw, err.Error(),
				http.StatusInternalServerError)
			return
		}
		// filename := h.Filename
		filename := time_now + ".jpg"
		defer f.Close()
 
		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		if err != nil {
			http.Error(rw, err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer t.Close()
 
		if _, err := io.Copy(t, f); err != nil {
			http.Error(rw, err.Error(),
				http.StatusInternalServerError)
			return
		}
		res.Data = map[string]string{"src": DOMAIN + "/img" + "?img=" + filename}
		res.Code = 0
		res.Message = "ok"
		res1, _ := json.Marshal(res)
		fmt.Fprint(rw, string(res1))
}

func imgdir(rw http.ResponseWriter, req *http.Request){
	 rw.Header().Set("Content-Type", "image/png")
	 req.ParseForm()
	if req.Form["img"] ==nil{
		res, _ := json.Marshal(Res{0,[]string{},"图片不存在"})
		fmt.Fprint(rw, string(res))
	    return
	}
	img :=req.Form["img"][0]
    file, err := ioutil.ReadFile(UPLOAD_DIR + "/" + img)
    if err != nil {
        fmt.Fprintf(rw,"查无此图片")
        return
    }
    rw.Write(file)
}

func cover(rw http.ResponseWriter, req *http.Request){
	rw.Header().Set("Access-Control-Allow-Origin", "*") 
	req.ParseForm()
	if req.Form["img"] ==nil || req.Form["title"] ==nil || req.Form["info"] ==nil  {
		res, _ := json.Marshal(Res{0,[]string{},"缺少参数"})
		fmt.Fprint(rw, string(res))
	    return
	}

	img :=req.Form["img"][0]
	title :=req.Form["title"][0]
	info :=req.Form["info"][0]

	now := time.Now().Unix()
	tm := time.Unix(now, 0)
	time_now :=tm.Format("2006-01-02")

	DB, _ := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test")
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)

	_,err:=DB.Exec("INSERT INTO cover(info,title,img,detail,addtime)VALUES (?,?,?,?,?)",info,title,img,"",time_now)

    if err != nil {
        res, _ := json.Marshal(Res{0,[]string{},"error"})
		fmt.Fprint(rw, string(res))
    }

    res1, _ := json.Marshal(Res{200,[]string{},"ok"})
	fmt.Fprint(rw, string(res1))
}

func list(rw http.ResponseWriter, req *http.Request){
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	req.ParseForm()

	page := "0"
	if req.Form["page"] != nil{
		page = req.Form["page"][0]
		if page == "1" {
			page = "0"
		}else{
			u, _ := strconv.Atoi(page)
			if u <= 0{
				fmt.Fprint(rw, "cuowu")
				return
			}
			num := (u - 1)*10
			page = strconv.Itoa(num)
		}
	}

	sel :="SELECT * FROM cover limit " + page +",10"
	DB, _ := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test")
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	rows,err:=DB.Query(sel)
	    //遍历打印
	    var s Info
	    s1 :=new(Res2)
	    for rows.Next(){
	        err=rows.Scan(&s.id,&s.info,&s.title,&s.img,&s.detail,&s.addtime)
	        s1.data = append(s1.data,s)
	    }
 		fmt.Println(s1)
	    if err != nil {
	        res, _ := json.Marshal(Res{0,[]string{},"error"})
			fmt.Fprint(rw, string(res))
	    }

	    //用完关闭
	    rows.Close()          
	    res, _ := json.Marshal(s1)
		fmt.Fprint(rw, string(res))


}
//md5 jiami 
func md5V(str string) string  {
    h := md5.New()
    h.Write([]byte(str))
    return hex.EncodeToString(h.Sum(nil))
}