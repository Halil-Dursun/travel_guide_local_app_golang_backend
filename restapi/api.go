package restapi

import (
	"encoding/json"
	"fmt"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

func Api() {
db, _ = sql.Open("mysql", "root:@tcp(localhost:3306)/goflut")

	router := mux.NewRouter()
	router.HandleFunc("/post", getPostsWithComments).Methods("GET")
	router.HandleFunc("/post/create", createPost).Methods("POST")
	router.HandleFunc("/post/update", updatePost).Methods("PUT")
	router.HandleFunc("/post/delete/{post_id}", deletePost).Methods("DELETE")
	router.HandleFunc("/post/{user_id}", getPostByUserID).Methods("GET")
	router.HandleFunc("/user", getUsers).Methods("GET")
	router.HandleFunc("/user/{id}", getUserById).Methods("GET")
	router.HandleFunc("/user/create", createUser).Methods("POST")
	router.HandleFunc("/user/update", updateUser).Methods("PUT")
	router.HandleFunc("/user/login", loginUser).Methods("POST")
	//router.HandleFunc("/comment/{post_id}", getCommentByPostId).Methods("GET")
	router.HandleFunc("/comment/create", createComments).Methods("POST")
	router.HandleFunc("/comment/delete/{comment_id}", deleteComment).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}

//! POST CRUD

//* Get posts from database, to api

func getPostsWithComments(w http.ResponseWriter, r *http.Request) {
	postList := GetPostsFromDatabase()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(postList)

}
func GetPostsFromDatabase() []map[string]interface{} {
	result, err := db.Query("SELECT * FROM post")
	checkErr(err)
	var postListWithComment []map[string]interface{}
	for result.Next() {
		var postModel Post
		result.Scan(&postModel.Id, &postModel.UserID, &postModel.Username, &postModel.Title, &postModel.Description, &postModel.City)
		resultComment, err := db.Query("SELECT * FROM comment WHERE post_id=?", postModel.Id)
		checkErr(err)
		var commentListByPost []Comment
		for resultComment.Next() {
			var commentModel Comment
			resultComment.Scan(&commentModel.ID, &commentModel.PostID, &commentModel.UserID, &commentModel.Username, &commentModel.Comment)
			commentListByPost = append(commentListByPost, commentModel)
		}
		mapItem := map[string]interface{}{
			"post":     postModel,
			"comments": commentListByPost}
		postListWithComment = append(postListWithComment, mapItem)
	}
	return postListWithComment
}


//* Create Post

func createPost(w http.ResponseWriter, r *http.Request) {
	var postModel Post
	err := json.NewDecoder(r.Body).Decode(&postModel)
	checkErr(err)
	w.WriteHeader(http.StatusOK)
	stmt, err := db.Prepare("INSERT INTO post set user_id=?, username=?,title=?,description=?,city=?")
	checkErr(err)
	_, err = stmt.Exec(postModel.UserID, postModel.Username, postModel.Title, postModel.Description, postModel.City)
	checkErr(err)
}

//* Update Post
func updatePost(w http.ResponseWriter, r *http.Request) {
	var postModel Post
	err := json.NewDecoder(r.Body).Decode(&postModel)
	checkErr(err)
	fmt.Print(postModel)
	stmt, err := db.Prepare("UPDATE post SET title=?, description=?,city=? WHERE id=?")
	checkErr(err)
	_, err = stmt.Exec(postModel.Title, postModel.Description, postModel.City, postModel.Id)
	checkErr(err)
}

//* Delete Post

func deletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["post_id"]
	_, err := db.Query("DELETE FROM post WHERE id=?", postID)
	checkErr(err)
	_, err = db.Query("DELETE FROM comment WHERE post_id =?", postID)
	checkErr(err)
}

//* Get Posts by user id

func getPostByUserID(w http.ResponseWriter, r *http.Request) {
	var userPostList []Post
	vars := mux.Vars(r)
	userID := vars["user_id"]
	result, err := db.Query("SELECT * FROM post WHERE user_id=?", userID)
	checkErr(err)
	for result.Next() {
		var postModel Post
		result.Scan(&postModel.Id, &postModel.UserID, &postModel.Username, &postModel.Title, &postModel.Description, &postModel.City)
		userPostList = append(userPostList, postModel)
	}
	err = json.NewEncoder(w).Encode(userPostList)
	checkErr(err)

}

//! USER CRUD

//* Get users from database,to api

func getUsers(w http.ResponseWriter, r *http.Request) {
	userList := getUsersFomDatabase()
	json.NewEncoder(w).Encode(userList)
	w.WriteHeader(http.StatusOK)
}

func getUsersFomDatabase() []User {
	result, err := db.Query("SELECT * FROM user")
	checkErr(err)

	var userList []User

	for result.Next() {
		var userModel User
		result.Scan(&userModel.ID, &userModel.Name, &userModel.Email, &userModel.Password)
		userList = append(userList, userModel)

	}
	return userList
}

//* Get User By Id

func getUserById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var userModel User
	userID := vars["id"]
	err := db.QueryRow("SELECT id,name,email FROM user WHERE id=?", userID).Scan(&userModel.ID,&userModel.Name, &userModel.Email)
	checkErr(err)
	err = json.NewEncoder(w).Encode(userModel)
	checkErr(err)
}

//* Login User

func loginUser(w http.ResponseWriter, r *http.Request) {
	var userModel User
	err := json.NewDecoder(r.Body).Decode(&userModel)
	checkErr(err)
	defer r.Body.Close()
	w.WriteHeader(http.StatusOK)
	var id int
	var email string
	var password string
	db.QueryRow("SELECT id,email,password FROM user WHERE email=?", userModel.Email).Scan(&id, &email, &password)
	resultMap := make(map[string]interface{})
	if id == 0 {
		resultMap["result"] = "no-account"
		json.NewEncoder(w).Encode(resultMap)
	} else {
		if userModel.Password == password {
			resultMap["result"] = id
			json.NewEncoder(w).Encode(resultMap)
		} else {
			resultMap["result"] = "wrong-password"
			json.NewEncoder(w).Encode(resultMap)
		}
	}
}

//* Create user

func createUser(w http.ResponseWriter, r *http.Request) {
	var userModel User
	err := json.NewDecoder(r.Body).Decode(&userModel)
	checkErr(err)
	defer r.Body.Close()
	w.WriteHeader(http.StatusOK)
	var id int
	db.QueryRow("SELECT id FROM user WHERE email=?", userModel.Email).Scan(&id)
	resultMap := make(map[string]interface{})
	if id != 0 {
		resultMap["result"] = "email-already-exist"
		print(id)
		json.NewEncoder(w).Encode(resultMap)
	} else {
		fmt.Println(id)
		db.QueryRow("INSERT INTO user set name=?, email=?,password=?", &userModel.Name, &userModel.Email, &userModel.Password)
		checkErr(err)
		var id int
		db.QueryRow("SELECT id FROM user WHERE email=?", &userModel.Email).Scan(&id)
		checkErr(err)
		resultMap["result"] = id
		json.NewEncoder(w).Encode(resultMap)
	}

}

//* Update User

func updateUser(w http.ResponseWriter, r *http.Request) {
	var userModel User
	err := json.NewDecoder(r.Body).Decode(&userModel)
	checkErr(err)
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)

	stmt, err := db.Prepare("UPDATE user SET name=?, email=?, password=? WHERE id=?")
	checkErr(err)
	_, err = stmt.Exec(userModel.Name, userModel.Email, userModel.Password, userModel.ID)
	checkErr(err)
}

//! COMMENT CRUD

//*Get comments by post id

// func getCommentByPostId(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	postID := vars["post_id"]
// 	var commentListByPostID []Comment

// 	result, err := db.Query("SELECT * FROM comment WHERE post_id=?", postID)
// 	checkErr(err)

// 	for result.Next() {
// 		var commentModel Comment
// 		result.Scan(&commentModel.ID, &commentModel.PostID, &commentModel.UserID, commentModel.Comment)
// 		commentListByPostID = append(commentListByPostID, commentModel)
// 	}

// 	err = json.NewEncoder(w).Encode(commentListByPostID)
// 	checkErr(err)

// }

//* Create Comment

func createComments(w http.ResponseWriter, r *http.Request) {
	var commentModel Comment
	err := json.NewDecoder(r.Body).Decode(&commentModel)
	checkErr(err)
	fmt.Println(commentModel)
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
	db.QueryRow("INSERT INTO comment set post_id=?, user_id=?,username=?,comment=?", &commentModel.PostID, &commentModel.UserID, &commentModel.Username,&commentModel.Comment)
	checkErr(err)
}

//* Delete Comment

func deleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["comment_id"]
	_, err := db.Query("DELETE from comment WHERE id=?", commentID)
	checkErr(err)
}

//! Sturcts

//* Post struct
type Post struct {
	Id          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Title       string `json:"title"`
	Description string `json:"description"`
	City        string `json:"city"`
}

//* User struct
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

//* Comment struct
type Comment struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	PostID   int    `json:"post_id"`
	Comment  string `json:"comment"`
}

//! Check Error

//* error check function
func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
