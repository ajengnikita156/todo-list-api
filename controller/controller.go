package controller

import (
	"database/sql"
	"fmt"
	"io"
	"membuatuser/helpers"
	"membuatuser/model"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

)

type MyClaims struct {
	jwt.StandardClaims
	ID int `json:"id"`
}

type Claims struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	jwt.StandardClaims
}

// menghitung jumlah status
func CountTexts(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		query := `
			SELECT 
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending,
				SUM(CASE WHEN status = 'progress' THEN 1 ELSE 0 END) AS progress,
				SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END) AS done
			FROM tasks 
			WHERE id_user = $1
		`

		counts := struct {
			Pending  int `json:"pending"`
			Progress int `json:"progress"`
			Done     int `json:"done"`
		}{}
		err := db.Get(&counts, query, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "successfully displays status data",
			"data":    counts,
		})
	}
}

// - API Search task (based on logged in user)
func SearchTasksFormController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var users []model.TaskRes
		var rows *sql.Rows
		claims := helpers.ClaimToken(c)
		id := claims.ID

		search := c.QueryParam("search")
		date := c.QueryParam("date")

		var parseDate time.Time
		if date != "" {
			layout := "2006-01-02"
			parseDate, err = time.Parse(layout, date)
			if err != nil {
				return err
			}
		}

		fmt.Println(date)

		query := `SELECT id, tittle, description, status, date, image, created_at, updated_at, id_user FROM tasks WHERE id_user = $1 AND (tittle ILIKE $2 OR description ILIKE $2)`

		search = "%" + search + "%"

		if !parseDate.IsZero() {
			query += "AND date::date = $3::date"
		}

		if !parseDate.IsZero() {
			rows, err = db.Query(query, id, search, parseDate)
		} else {
			rows, err = db.Query(query, id, search)
		}

		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var user model.TaskRes
			err = rows.Scan(
				&user.ID,
				&user.Tittle,
				&user.Description,
				&user.Status,
				&user.Date,
				&user.Image,
				&user.CreatedAt,
				&user.UpdatedAt,
				&user.IdUser,
			)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		if len(users) == 0 {
			users = []model.TaskRes{}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Success Search Tasks for User",
			"data":    users,
		})
	}
}

// - API Show task list (based on logged in user)
func GetTasksController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var users []model.TaskRes
		claims := helpers.ClaimToken(c)
		id := claims.ID
		fmt.Println(id)

		query := `SELECT id, tittle, description, status, date, image, created_at, updated_at, id_user FROM tasks WHERE id_user = $1`

		rows, err := db.Query(query, id)
		if err != nil {
			return err
		}

		for rows.Next() {
			var user model.TaskRes
			err = rows.Scan(
				&user.ID,
				&user.Tittle,
				&user.Description,
				&user.Status,
				&user.Date,
				&user.Image,
				&user.CreatedAt,
				&user.UpdatedAt,
				&user.IdUser,
			)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Success Get Data Task By User Login",
			"data":    users,
		})
	}
}

// - API Show detail task
func GetTaskByIDController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var task model.TaskRes
		claims := helpers.ClaimToken(c)
		id := claims.ID
		taskId := c.Param("id")

		query := `SELECT id, tittle, description, status, date, image, created_at, updated_at, id_user FROM tasks WHERE id_user = $1 AND id = $2`

		rows, err := db.Query(query, id, taskId)
		if err != nil {
			return err
		}

		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(
				&task.ID,
				&task.Tittle,
				&task.Description,
				&task.Status,
				&task.Date,
				&task.Image,
				&task.CreatedAt,
				&task.UpdatedAt,
				&task.IdUser,
			)
			if err != nil {
				return err
			}
		} else {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "Task Not Found",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Success Get Task Detail",
			"data":    task,
		})
	}
}

// EDIT TASK -  - API Update a task
func EditTaskController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		taskID := c.Param("id")
		var request model.TaskReq
		var user model.TaskRes
		err := c.Bind(&request)

		if err != nil {
			return err
		}

		layout := "2006-01-02 15:04:05"
		parseDate, err := time.Parse(layout, request.Date)
		if err != nil {
			return err
		}

		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		validate := validator.New()
		err = validate.Struct(request)
		if err != nil {
			var errormessages []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errormessages = append(errormessages, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errormessages,
			})
		}

		query := `UPDATE tasks SET tittle = $1, description = $2, status = $3, date = $4,
				id_user = $5, updated_at = now() WHERE id = $6
				RETURNING id, tittle, description, status, date, image, created_at, updated_at, id_user`

		row := db.QueryRowx(query, request.Tittle, request.Description, request.Status, parseDate, id, taskID)
		err = row.Scan(
			&user.ID,
			&user.Tittle,
			&user.Description,
			&user.Status,
			&user.Date,
			&user.Image,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IdUser,
		)

		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Successfully edited Data ",
			"data":    user,
		})
	}
}

// ADD TASK - API Create a new task
func AddTaskController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.TaskReq
		var user model.TaskRes
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		// Menerima file gambar dari form dengan nama "image"
		image, err := c.FormFile("image")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Tidak dapat memproses file gambar"})
		}

		// Buka file yang diunggah
		src, err := image.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal membuka file gambar"})
		}
		defer src.Close()

		// Lokasi penyimpanan file gambar lokal
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal membuat direktori penyimpanan"})
		}

		// Generate nama file unik
		fileName := fmt.Sprintf("%s%s", image.Filename, filepath.Ext(image.Filename))
		dstPath := filepath.Join(uploadDir, fileName)

		// Membuka file tujuan untuk penyimpanan
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal membuat file gambar"})
		}
		defer dst.Close()

		// Salin isi file dari file asal ke file tujuan
		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyalin file gambar"})
		}

		// Membuat URL ke gambar yang diunggah
		imageURL := "http://localhost:8090/" + dstPath

		layout := "2006-01-02 15:04"
		parsedDate, err := time.Parse(layout, req.Date)
		if err != nil {
			return err
		}

		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		validate := validator.New()
		err = validate.Struct(req)
		if err != nil {
			var errorMessage []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errorMessage = append(errorMessage, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errorMessage,
			})
		}

		query := `
		INSERT INTO tasks (tittle, description, status, date, image, created_at, id_user)
		VALUES ($1, $2, $3, $4, $5, now(), $6)  
		RETURNING id, tittle, description, status, date, image, created_at, updated_at, id_user
		`
		row := db.QueryRowx(query, req.Tittle, req.Description, req.Status, parsedDate, imageURL, id)
		err = row.Scan(&user.ID, &user.Tittle, &user.Description, &user.Status, &user.Date, &user.Image, &user.CreatedAt, &user.UpdatedAt, &user.IdUser)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Successfully Added New TASK Data",
			"data":    user,
		})
	}
}

//   - API Delete a task and bulk delete   AND id_user = $2
//
// delete
func DeleteTaskController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		taskID := c.Param("id")

		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		query := "DELETE FROM tasks WHERE id = $1 AND id_user = $2"
		_, err := db.Exec(query, taskID, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Task Data Deleted Successfully",
		})
	}
}

// bulk delete
func BulkDeleteController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request model.BulkDeleteRequest
		err := c.Bind(&request)

		for _, id := range request.ID {
			query := "DELETE FROM tasks WHERE id = $1"
			_, err = db.Exec(query, id)
			if err != nil {
				return err
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "successful deletion of some data",
		})
	}
}

//- API Authentication (login, register and logout)

// REGISTRASI
func RegisterController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var register model.UserRegis
		var user model.UserRegisRespon
		validate := validator.New()

		err := c.Bind(&register) //=>mencocokan data di struct
		//bind =>mengambil data dari input,mengisi var register ,dan mencocokan di struct userreq
		if err != nil {
			return err
		}

		//function validator
		err = validate.Struct(register)
		if err != nil {
			var errormassage []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errormassage = append(errormassage, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errormassage,
			})
		}

		//function hasspassword
		password, err := helpers.HashPassword(register.Password)
		if err != nil {
			return err
		}

		query := `
		INSERT INTO users (email, password, created_at)
		VALUES ( $1, $2, now())
		RETURNING id, email, created_at `

		row := db.QueryRowx(query, register.Email, password) //=>mengambil data yang sama di struct
		err = row.Scan(
			&user.ID,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Successfully Registered",
			"data":    user,
		})
	}
}

// LOGIN
func LoginCompareController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var login model.UserLogin
		var user model.UserLogRespon
		validate := validator.New()

		err := c.Bind(&login) //=>mencocokan data di struct

		if err != nil {
			return err
		}

		//function validator
		err = validate.Struct(login)
		if err != nil {
			var errormassage []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errormassage = append(errormassage, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errormassage,
			})
		}

		query := `SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1 `

		row := db.QueryRowx(query, login.Email) //=>mengambil data yang sama di struct
		err = row.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Email not registered",
				})
			}
			return err
		}

		match, err := helpers.ComparePassword(user.Password, login.Password)

		if err != nil {
			if !match {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{

					"message": "Passwords do not match",
				})
			}
			return err
		}

		var (
			jwtToken  *jwt.Token
			secretKey = []byte("secret")
		)

		jwtClaims := &Claims{
			ID:    user.ID,
			Email: user.Email,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			},
		}

		jwtToken = jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

		token, err := jwtToken.SignedString(secretKey)
		if err != nil {
			return err
		}

		const query2 = `INSERT INTO user_token (user_id, token) VALUES ($1, $2)`
		_ = db.QueryRowx(query2, user.ID, token)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Login Successful",
			"token":   token,
			"data":    user,
		})
	}
}

// LOGOUT
func LogoutController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var reqToken string
		headerDataToken := c.Request().Header.Get("Authorization")

		splitToken := strings.Split(headerDataToken, "Bearer ")
		if len(splitToken) > 1 {
			reqToken = splitToken[1]
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		query := `DELETE FROM user_token WHERE token = $1`

		_, err := db.Exec(query, reqToken)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]interface{}{
					"message": "Data pengguna tidak ditemukan",
				})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Succesfully Logout",
		})
	}

}
