package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ebrym/bookapi/data"
	"github.com/ebrym/bookapi/service"
	"github.com/ebrym/bookapi/utils"
	"github.com/gorilla/mux"
)

// RefreshToken handles refresh token request
func (ah *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	user := r.Context().Value(UserKey{}).(data.User)
	accessToken, err := ah.authService.GenerateAccessToken(&user)
	if err != nil {
		ah.logger.Error("unable to generate access token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		// data.ToJSON(&GenericError{Error: err.Error()}, w)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to generate access token.Please try again later"}, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	// data.ToJSON(&TokenResponse{AccessToken: accessToken}, w)
	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Successfully generated new access token",
		Data:    &TokenResponse{AccessToken: accessToken},
	}, w)
}

// Greet request greet request
func (ah *AuthHandler) Greet(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(UserIDKey{}).(string)
	w.WriteHeader(http.StatusOK)
	// w.Write([]byte("hello, " + userID))
	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "hello," + userID,
	}, w)
}

// GeneratePassResetCode generate a new secret code to reset password.
func (ah *AuthHandler) GeneratePassResetCode(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(UserIDKey{}).(string)

	user, err := ah.repo.GetUserByID(context.Background(), userID)
	if err != nil {
		ah.logger.Error("unable to get user to generate secret code for password reset", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to send password reset code. Please try again later"}, w)
		return
	}

	// Send verification mail
	from := "ibrodex@gmail.com"
	to := []string{user.Email}
	subject := "Password Reset for Bookite"
	mailType := service.PassReset
	mailData := &service.MailData{
		Username: user.Username,
		Code:     utils.GenerateRandomString(8),
	}

	mailReq := ah.mailService.NewMail(from, to, subject, mailType, mailData)
	err = ah.mailService.SendMail(mailReq)
	if err != nil {
		ah.logger.Error("unable to send mail", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to send password reset code. Please try again later"}, w)
		return
	}

	// store the password reset code to db
	verificationData := &data.VerificationData{
		Email:     user.Email,
		Code:      mailData.Code,
		Type:      data.PassReset,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(ah.configs.PassResetCodeExpiration)),
	}

	err = ah.repo.StoreVerificationData(context.Background(), verificationData)
	if err != nil {
		ah.logger.Error("unable to store password reset verification data", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to send password reset code. Please try again later"}, w)
		return
	}

	ah.logger.Debug("successfully mailed password reset code")
	w.WriteHeader(http.StatusOK)
	data.ToJSON(&GenericResponse{Status: true, Message: "Please check your mail for password reset code"}, w)
}

// get category request
func (cat *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	categoryList, err := cat.repo.GetCategories(context.Background())
	if err != nil {
		cat.logger.Error("unable to get categories", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "unable to get categories. Please try again later"}, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	catList := []CategoryUpdate{}
	for c := range categoryList {
		cat := CategoryUpdate{}

		cat.Id = categoryList[c].ID
		cat.Name = categoryList[c].Name
		cat.Code = categoryList[c].Code

		catList = append(catList, cat)
	}

	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Success",
		Data:    catList, //&CategoryUpdate{Id: category.ID, Code: category.Code, Name: category.Name},
	}, w)
}

// get category request
func (cat *CategoryHandler) GetCategoryById(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	categoryID := mux.Vars(r)["id"] //r.Context().Value(CategoryIDKey{}).(string)
	//ategoryID := r.URL.Query().Get("id")
	cat.logger.Debug("querying for category with Code", categoryID)
	category, err := cat.repo.GetCategoryByID(context.Background(), categoryID)
	if err != nil {
		cat.logger.Error("unable to get category", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "unable to get category. Please try again later"}, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	// w.Write([]byte("hello, " + userID))
	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Success",
		Data:    &CategoryUpdate{Id: category.ID, Code: category.Code, Name: category.Name},
	}, w)
} // get category by code request
func (cat *CategoryHandler) GetCategoryByCode(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	categoryID := mux.Vars(r)["code"]
	category, err := cat.repo.GetCategoryByCode(context.Background(), categoryID)
	if err != nil {
		cat.logger.Error("unable to get category", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "unable to get category. Please try again later"}, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	// w.Write([]byte("hello, " + userID))
	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Success",
		Data:    &CategoryUpdate{Id: category.ID, Code: category.Code, Name: category.Name},
	}, w)
}
