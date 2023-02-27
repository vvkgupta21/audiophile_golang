package handler

import (
	"audio_phile/database"
	"audio_phile/database/dbHelper"
	"audio_phile/middleware"
	"audio_phile/model"
	"audio_phile/utils"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var body model.UserRequestBody

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	parseBody := model.UserRequestBody{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
		return
	}

	if len(body.Password) < 6 {
		utils.RespondError(w, http.StatusBadRequest, nil, "password must be 6 chars long")
		return
	}

	exist, existErr := dbHelper.IsUserExist(body.Email)
	if exist {
		utils.RespondError(w, http.StatusBadRequest, existErr, "User already exist")
		return
	}

	if existErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existErr, "Failed to check existence")
		return
	}

	hashPassword, hasErr := utils.HashPassword(body.Password)

	if hasErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, hasErr, "failed to secure password")
		return
	}

	var userID string
	var err error

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userID, err = dbHelper.CreateUser(tx, body.Name, body.Email, hashPassword)

		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create user")
			return err
		}

		roleErr := dbHelper.CreateUserRole(tx, userID, model.RoleUser)
		if roleErr != nil {
			return roleErr
		}
		return nil
	})
	// error message correct karo
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to create user")
		return
	}

	//code could be 201
	utils.RespondJSON(w, http.StatusCreated, model.UserResponseBody{
		UserId: userID,
		Name:   body.Name,
		Email:  body.Email,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body model.LoginRequestBody

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	parseBody := model.LoginRequestBody{
		Email:    body.Email,
		Password: body.Password,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
		return
	}

	userId, err := dbHelper.GetUserIDByEmailAndPassword(body.Email, body.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, http.StatusBadRequest, errors.New("user does not exist"), "user does not exist")
			return
		}
		utils.RespondError(w, http.StatusBadRequest, err, "Incorrect credentials")
		return
	}

	role, err := dbHelper.GetUserRoles(userId)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error in getting user role")
		return
	}

	token, err := middleware.GenerateJWT(userId, role)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error in generating jwt token")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body model.ProductsRequest

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	parseBody := model.ProductsRequest{
		Name:        body.Name,
		Price:       body.Price,
		Description: body.Description,
		IsAvailable: body.IsAvailable,
		Quantity:    body.Quantity,
		Category:    body.Category,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
		return
	}

	exist, existErr := dbHelper.IsProductExist(body.Name)
	if exist {
		utils.RespondError(w, http.StatusBadRequest, existErr, "Product already exist")
		return
	}

	if existErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existErr, "Failed to product existence")
		return
	}

	productId, err := dbHelper.CreateProduct(database.Audiophile, body.Name, body.Description, body.IsAvailable, body.Price, body.Quantity, body.Category)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create product")
		return
	}

	//code could be 201
	utils.RespondJSON(w, http.StatusCreated, model.ProductsResponse{
		ProductId:   productId,
		Name:        body.Name,
		Price:       body.Price,
		Description: body.Description,
		IsAvailable: body.IsAvailable,
		Quantity:    body.Quantity,
		Category:    body.Category,
	})
}

func GetAllProduct(w http.ResponseWriter, r *http.Request) {
	list, err := dbHelper.GetAllProduct()
	logrus.Println(list)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, list)
	if err != nil {
		return
	}
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	productId := chi.URLParam(r, "id")
	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		return
	}
	utils.RespondJSON(w, http.StatusOK, productDetail)
}

func CreatedAddress(w http.ResponseWriter, r *http.Request) {
	// parse the request data
	var addresses []model.AddressRequest
	if err := utils.ParseBody(r.Body, &addresses); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	userId := getUserId(r)
	for _, address := range addresses {
		// save the address to the database
		err := dbHelper.CreateAddresses(database.Audiophile, userId, address.Address, address.AddressType, address.Lat, address.Long)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create role")
			return
		}
	}
	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string
	}{"Address created successfully!!"})
}

func GetUserByUserId(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	var userData model.User
	var userAddress []model.AddressModel
	var userDetails model.UserWithAddress
	var err error
	userData, err = dbHelper.GetUserByUserId(userId)
	if err != nil {
		return
	}
	userAddress, err = dbHelper.GetAddress(database.Audiophile, userId)
	if err != nil {
		return
	}
	userDetails.Id = userId
	userDetails.Name = userData.Name
	userDetails.Email = userData.Email
	userDetails.Address = userAddress
	utils.RespondJSON(w, http.StatusOK, userDetails)
}

func getUserId(r *http.Request) string {
	user := r.Context().Value(middleware.UserContext).(map[string]interface{})
	fmt.Println(user)
	var userId string
	userId = user["id"].(string)
	fmt.Println(userId)
	return userId
}

func GetAllUser(w http.ResponseWriter, r *http.Request) {
	list, err := dbHelper.GetAllUser(model.RoleUser)
	logrus.Println(list)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, list)
	if err != nil {
		return
	}
}

func DeleteUserByUserId(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	err := dbHelper.DeleteUser(database.Audiophile, userId)
	if err != nil {
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"User deleted successfully!"})
}

func CreateProductToCart(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(r)
	productId := chi.URLParam(r, "id")
	quantityStr := chi.URLParam(r, "quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to covert from string to int")
		return
	}

	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Product not found!")
		return
	}

	if quantity > productDetail.Quantity {
		utils.RespondError(w, http.StatusBadRequest, nil, "Requested quantity not available")
		return
	}

	existingCartId, exist, existErr := dbHelper.IsCartExist(userId, model.CartStatusActive)

	if existErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existErr, "Failed to check cart existence")
		return
	}

	if exist {
		err := dbHelper.CreateProductInCart(database.Audiophile, existingCartId, productId, quantity)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to add product")
			return
		}
	} else {
		txErr := database.Tx(func(tx *sqlx.Tx) error {
			cartId, err := dbHelper.CreateCart(tx, userId, model.CartStatusActive)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create cart")
				return err
			}

			err = dbHelper.CreateProductInCart(tx, cartId, productId, quantity)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Failed to add product")
				return err
			}
			return nil
		})
		// error message correct karo
		if txErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, txErr, "transaction error")
			return
		}
	}

	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string
	}{"Product added to cart!"})
}

func GetCartWithProductById(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(r)
	list, err := dbHelper.GetCartWithProduct(database.Audiophile, userId)
	logrus.Println(list)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, list)
	if err != nil {
		return
	}
}

func AddProductQuantityInCart(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	productId := chi.URLParam(r, "productId")

	cartDetail, err := dbHelper.GetCartProductQuantity(cartProductId, productId)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Cart detail not found!")
		return
	}

	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Product not found!")
		return
	}

	if cartDetail.Quantity >= productDetail.Quantity {
		utils.RespondError(w, http.StatusBadRequest, nil, "Requested quantity not available")
		return
	}

	err = dbHelper.AddProductQuantityInCart(cartProductId, productId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"Quantity updated successfully"})
}

func RemoveProductQuantityInCart(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	productId := chi.URLParam(r, "productId")

	cartDetail, err := dbHelper.GetCartProductQuantity(cartProductId, productId)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Cart detail not found!")
		return
	}

	if cartDetail.Quantity <= 0 {
		utils.RespondError(w, http.StatusBadRequest, nil, "Requested quantity not available")
		return
	}

	err = dbHelper.RemoveProductQuantityInCart(cartProductId, productId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"Quantity updated successfully"})

}

func DeleteProductFromCart(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	productId := chi.URLParam(r, "productId")

	err := dbHelper.DeleteProductFromCart(cartProductId, productId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"Product deleted successfully"})
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	orderId, err := dbHelper.CreateOrder(cartProductId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
		OrderId string
	}{Message: "Order placed successfully", OrderId: orderId})
}
