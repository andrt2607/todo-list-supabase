package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"todo-list/db"
	"todo-list/models"
	"todo-list/utils"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/xuri/excelize/v2"
)

// func Login(c *gin.Context) {
// 	type LoginData struct {
// 		Username string `json:"username" binding:"required"`
// 		Password string `json:"password" binding:"required"`
// 	}
// 	var loginData LoginData
// 	var user models.UserModel

// 	err := c.ShouldBindJSON(&loginData)
// 	if err != nil {
// 		utils.ThrowErr(c, http.StatusBadRequest, "Invalid Request")
// 		return
// 	}

// 	if err = db.DB.QueryRow("select id, username, password, is_active from users where username = $1", loginData.Username).Scan(&user.Id, &user.Username, &user.Password, &user.IsActive); err != nil {
// 		if err == sql.ErrNoRows {
// 			utils.ThrowErr(c, http.StatusUnauthorized, "Username or Password is wrongd")
// 		} else {
// 			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
// 		}
// 		return
// 	}

// 	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
// 		utils.ThrowErr(c, http.StatusUnauthorized, "Username or Password is wrong")
// 		return
// 	}

// 	if !user.IsActive {
// 		utils.ThrowErr(c, http.StatusUnauthorized, "Your account is not active")
// 		return
// 	}

// 	sign := jwt.New(jwt.SigningMethodHS256)
// 	claims := sign.Claims.(jwt.MapClaims)
// 	claims["userId"] = user.Id
// 	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // 30 days\

// 	token, err := sign.SignedString([]byte(os.Getenv("JWT_SECRET")))
// 	if err != nil {
// 		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"errors":  false,
// 		"message": "Login Success",
// 		"data": gin.H{
// 			"token":   token,
// 			"type":    "Bearer",
// 			"expired": "30d",
// 		},
// 	})
// }

// in below code, we will create a new controller for activities with validator as a dependency
type ActivityController struct {
	validator *validator.Validate
}

// NewActivityController is a constructor to create a new activity controller
func NewActivityController(validator *validator.Validate) *ActivityController {
	return &ActivityController{validator}
}

func (ac *ActivityController) GetActivities(c *gin.Context) {
	rows, err := db.DB.Query("select * from activities")
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusUnauthorized, "Username or Password is wrongd")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	var activities []models.ActivityModel
	for rows.Next() {
		var activity models.ActivityModel
		err = rows.Scan(&activity.Id, &activity.Title, &activity.CreatedAt, &activity.Description, &activity.Status, &activity.Category)
		if err != nil {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
			return
		}
		activities = append(activities, activity)
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Get Activities Success",
		"data":    activities,
	})
}

func (ac *ActivityController) CreateActivity(c *gin.Context) {
	var activity models.ActivityModel
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	if err := ac.validator.Struct(activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	var id int
	sqlStatement := `INSERT INTO activities(title, category, description, status) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.DB.QueryRow(sqlStatement, activity.Title, activity.Category, activity.Description, activity.Status).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to create activity",
		})
		return
	}

	activity.Id = id
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Activity Created Successfully",
		"data":    activity,
	})
}

func (ac *ActivityController) UpdateActivity(c *gin.Context) {
	idInput := c.Param("id")
	var activity models.ActivityModel
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	if err := ac.validator.Struct(activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	sqlStatement := `UPDATE activities SET title=$1, category=$2, description=$3, status=$4 WHERE id=$5`
	_, err := db.DB.Exec(sqlStatement, activity.Title, activity.Category, activity.Description, activity.Status, idInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update activity",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Activity Updated Successfully",
	})
}

func (ac *ActivityController) DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	sqlStatement := `DELETE FROM activities WHERE id=$1`
	_, err := db.DB.Exec(sqlStatement, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete activity",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Activity Deleted Successfully",
	})
}

func (ac *ActivityController) ExportExcelActivities(c *gin.Context) {
	var activities []models.ActivityModel
	rows, err := db.DB.Query("select * from activities")
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusUnauthorized, "Username or Password is wrongd")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	for rows.Next() {
		var activity models.ActivityModel
		err = rows.Scan(&activity.Id, &activity.Title, &activity.CreatedAt, &activity.Description, &activity.Status, &activity.Category)
		if err != nil {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
			return
		}
		activities = append(activities, activity)
	}

	fileExcel := excelize.NewFile()
	defer func() {
		if err := fileExcel.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	index, err := fileExcel.NewSheet("Activities")
	if err != nil {
		fmt.Println(err)
		return
	}
	fileExcel.SetCellValue("Activities", "A1", "ID")
	fileExcel.SetCellValue("Activities", "B1", "Title")
	fileExcel.SetCellValue("Activities", "C1", "Created At")
	fileExcel.SetCellValue("Activities", "D1", "Description")
	fileExcel.SetCellValue("Activities", "E1", "Status")
	fileExcel.SetCellValue("Activities", "F1", "Category")
	// i want custom width for each column
	fileExcel.SetColWidth("Activities", "A", "A", 10)
	fileExcel.SetColWidth("Activities", "B", "B", 20)
	fileExcel.SetColWidth("Activities", "C", "C", 20)
	fileExcel.SetColWidth("Activities", "D", "D", 40)
	fileExcel.SetColWidth("Activities", "E", "E", 30)
	fileExcel.SetColWidth("Activities", "F", "F", 20)
	// i want custom styling with border
	style, err := fileExcel.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	err = fileExcel.SetCellStyle("Activities", "A1", "F1", style)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Define a date style for column C
	dateStyle, err := fileExcel.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		NumFmt:    22, // date format
	})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	for i, activity := range activities {
		parseTime, err := time.Parse(time.RFC3339, activity.CreatedAt)
		if err != nil {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
			return
		}
		indonesiaParseTime := parseTime.In(time.FixedZone("WIB", 7*3600))
		fileExcel.SetCellValue("Activities", fmt.Sprintf("A%d", i+2), activity.Id)
		fileExcel.SetCellValue("Activities", fmt.Sprintf("B%d", i+2), activity.Title)
		fileExcel.SetCellValue("Activities", fmt.Sprintf("C%d", i+2), indonesiaParseTime)
		fileExcel.SetCellValue("Activities", fmt.Sprintf("D%d", i+2), activity.Description)
		fileExcel.SetCellValue("Activities", fmt.Sprintf("E%d", i+2), activity.Status)
		fileExcel.SetCellValue("Activities", fmt.Sprintf("F%d", i+2), activity.Category)
		err = fileExcel.SetCellStyle("Activities", fmt.Sprintf("A%d", i+2), fmt.Sprintf("F%d", i+2), style)
		if err != nil {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
			return
		}
		err = fileExcel.SetCellStyle("Activities", fmt.Sprintf("C%d", i+2), fmt.Sprintf("C%d", i+2), dateStyle)
		if err != nil {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	fileExcel.SetActiveSheet(index)
	if err := fileExcel.SaveAs("excel/Activities.xlsx"); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	contentDisposition := fmt.Sprintf("attachment; filename=%s", "Activities.xlsx")
	c.Writer.Header().Set("Content-Disposition", contentDisposition)
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.File("excel/Activities.xlsx")
	//delete after download
	if err := os.Remove("excel/Activities.xlsx"); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (ac *ActivityController) ExportPDFActivities(c *gin.Context) {
	// exec.Command("mkdir", "-p", "pdf")
	// exec.Command("rm", "-rf", "pdf/output_activity.pdf")
	// command := exec.Command("wkhtmltopdf", "template_html/activity_template.html", "pdf/output_activity.pdf")
	// err := command.Run()
	// if err != nil {
	// 	utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
	// 	return
	// }
	// const path = "C:/Program Files/wkhtmltopdf/bin"
	// wkhtmltopdf.SetPath(path)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	templateHtml, err := os.Open("template_html/activity_template.html")
	if templateHtml != nil {
		defer templateHtml.Close()
	}
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	pdfg.AddPage(wkhtmltopdf.NewPageReader(templateHtml))
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(300)
	err = pdfg.Create()
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	err = pdfg.WriteFile("./pdf/output_activity.pdf")
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
}
